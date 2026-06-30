// Rescue Dispatch Score (trackd-rescue-dispatch, plan.md §21.4).
//
// Moves Return Rescue from "nearby feed sorted by recency" to Uber/FoodMatch-
// style local dispatch: a transparent weighted utility over each (unit, buyer)
// edge. We do not need full RL — a clear, explainable score is enough for MVP.
//
//	dispatch_score =
//	    w_demand   · demand_intent
//	  + w_distance · distance_savings
//	  + w_ttl      · ttl_urgency
//	  + w_price    · price_acceptance
//	  + w_keep     · buyer_keep_probability
//	  + w_carbon   · carbon_saved
//	  - w_failrisk · failed_claim_risk
//	  - w_chainrisk· chain_depth_risk
//
// relay-api precomputes every signal; this stays a pure decisioning layer.
package scoring

import (
	"github.com/shikherjha/relay-engine/internal/carbon"
	"github.com/shikherjha/relay-engine/internal/config"
	"github.com/shikherjha/relay-engine/internal/models"
)

// DispatchWeights are the tunable term weights (positives ideally sum to 1.0;
// the two risk terms subtract). Sourced from config so they stay in sync with
// the relay-api mock (app/clients/engine_client.py).
type DispatchWeights struct {
	Demand    float64
	Distance  float64
	TTL       float64
	Price     float64
	Keep      float64
	Carbon    float64
	FailRisk  float64
	ChainRisk float64
}

// DispatchWeightsFromConfig reads the weights from runtime config.
func DispatchWeightsFromConfig(cfg config.Config) DispatchWeights {
	return DispatchWeights{
		Demand:    cfg.DispatchWDemand,
		Distance:  cfg.DispatchWDistance,
		TTL:       cfg.DispatchWTTL,
		Price:     cfg.DispatchWPrice,
		Keep:      cfg.DispatchWKeep,
		Carbon:    cfg.DispatchWCarbon,
		FailRisk:  cfg.DispatchWFailRisk,
		ChainRisk: cfg.DispatchWChainRisk,
	}
}

const (
	dispatchCarbonNormKg = 3.0  // kg net CO2 mapped to a full carbon term
	dispatchWishFloor    = 0.45 // min viewer wish-match to read "matches your wish"
	dispatchMaxDiscount  = 0.45 // matches rescue_discount_max (price-depth scale)
	dispatchReturnCap    = 0.4  // high-return-rate risk threshold (rescue_user_return_rate_cap)
)

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func saturate(x float64) float64 { return x / (x + 1.0) } // 0..1 diminishing

// ScoreDispatch ranks each candidate edge and attaches explainable reasons.
func ScoreDispatch(req models.DispatchRequest, w DispatchWeights) models.DispatchResponse {
	scores := make([]models.DispatchScore, 0, len(req.Candidates))
	for _, c := range req.Candidates {
		s, reasons := scoreCandidate(req.Viewer, c, w)
		scores = append(scores, models.DispatchScore{
			ListingID:       c.ListingID,
			DispatchScore:   round3(s),
			DispatchReasons: reasons,
		})
	}
	return models.DispatchResponse{Scores: scores}
}

func scoreCandidate(v models.DispatchViewer, c models.DispatchCandidate, w DispatchWeights) (float64, []models.DispatchReason) {
	// demand_intent — open-wish demand near the unit, lifted by the viewer's own match.
	demandScore := 0.0
	if c.Demand != nil {
		demandScore = c.Demand.DemandScore
	}
	demandIntent := clamp01(saturate(demandScore))
	if c.ViewerWishMatch > demandIntent {
		demandIntent = clamp01(c.ViewerWishMatch)
	}

	// distance_savings — a closer local pickup saves more last-mile carbon.
	distanceSavings := 0.1 // ships (national) → minimal local saving
	if c.DistanceKm != nil {
		radius := c.RadiusKm
		if radius <= 0 {
			radius = 15.0
		}
		distanceSavings = clamp01(1.0 - *c.DistanceKm/radius)
	}

	// ttl_urgency — a decaying local listing nearing expiry should clear now.
	ttlUrgency := 0.0
	if c.TtlRemainingFrac != nil {
		ttlUrgency = clamp01(1.0 - *c.TtlRemainingFrac)
	}

	// price_acceptance — snug budget fit wins; else markdown depth.
	priceAcceptance := clamp01(c.DiscountPct/dispatchMaxDiscount) * 0.8
	if c.PriceFit {
		priceAcceptance = 1.0
	}

	// buyer_keep_probability — good grade + right size ⇒ unlikely to re-return.
	keep := c.GradeNumeric
	if !c.SizeFit {
		keep *= 0.85
	}
	keep = clamp01(keep)

	// carbon_saved — channel net of last-mile, normalized.
	channel := c.Channel
	if channel == "" {
		channel = "rescue"
	}
	carbonNorm := clamp01(carbon.NetSaved(channel, c.DeliveryKm) / dispatchCarbonNormKg)

	// failed_claim_risk — eligibility / high historical return rate.
	failRisk := 0.0
	if !v.Eligible {
		failRisk = 1.0
	}
	if v.ReturnRate >= dispatchReturnCap && failRisk < 0.6 {
		failRisk = 0.6
	}

	// chain_depth_risk — diminishing returns recirculating a well-travelled unit.
	chainRisk := clamp01(float64(c.TransferCount) / float64(chainDepthCap))

	score := w.Demand*demandIntent + w.Distance*distanceSavings + w.TTL*ttlUrgency +
		w.Price*priceAcceptance + w.Keep*keep + w.Carbon*carbonNorm -
		w.FailRisk*failRisk - w.ChainRisk*chainRisk

	reasons := dispatchReasons(c, demandIntent, distanceSavings, ttlUrgency, priceAcceptance, keep, carbonNorm, failRisk, chainRisk)
	return clamp01(score), reasons
}

// dispatchReasons builds the explainable "why you're seeing this" chips, capped
// at three so a card stays scannable. Positives lead, but any triggered risk
// chip reserves a slot so a guardrail caveat is never crowded out.
func dispatchReasons(c models.DispatchCandidate, demand, dist, ttl, price, keep, carbonN, failRisk, chainRisk float64) []models.DispatchReason {
	type cand struct {
		ok          bool
		code, label string
	}
	positives := []cand{
		{c.ViewerWishMatch >= dispatchWishFloor, "matches_your_wish", "Matches your wish"},
		{c.DistanceKm != nil && dist >= 0.55 && (demand >= 0.45 || c.ViewerWishMatch >= 0.3), "best_local_fit", "Best local fit"},
		{ttl >= 0.6, "ttl_urgent", "Clearing soon"},
		{c.PriceFit, "price_fit", "In your budget"},
		{!c.PriceFit && price >= 0.6, "priced_to_clear", "Priced to clear"},
		{carbonN >= 0.6, "high_carbon_save", "High carbon save"},
		{keep >= 0.85, "high_keep", "Great condition"},
	}
	risks := []cand{
		{failRisk >= 0.6, "claim_risk", "Eligibility limits this"},
		{chainRisk >= 0.66, "chain_depth", "Near reuse limit"},
	}
	riskOut := make([]models.DispatchReason, 0, 2)
	for _, a := range risks {
		if a.ok {
			riskOut = append(riskOut, models.DispatchReason{Code: a.code, Label: a.label})
		}
	}
	posBudget := 3 - len(riskOut)
	if posBudget < 0 {
		posBudget = 0
	}
	out := make([]models.DispatchReason, 0, 3)
	for _, a := range positives {
		if a.ok {
			out = append(out, models.DispatchReason{Code: a.code, Label: a.label})
			if len(out) >= posBudget {
				break
			}
		}
	}
	out = append(out, riskOut...)
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}
