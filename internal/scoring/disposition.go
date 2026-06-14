// Package scoring is the authoritative disposition rule engine (engine-disposition).
// Rules are the floor; demand re-weights among viable channels (engine-demand-weight).
// Guardrails: chain-depth cap, net-carbon gate, exchange-first.
package scoring

import (
	"fmt"

	"github.com/shikherjha/relay-engine/internal/carbon"
	"github.com/shikherjha/relay-engine/internal/models"
)

const (
	chainDepthCap   = 3
	gradeGoodFloor  = 0.6 // >= -> resale-grade (rescue/p2p)
	gradeFairFloor  = 0.3 // >= -> refurb; below -> donate/recycle
	demandWeight    = 0.3
	defaultDelivery = 5.0 // km, when no nearest-wish distance is supplied
)

var sizeReasons = map[string]bool{"too_small": true, "too_large": true, "fit": true}

// Score routes a returned unit to its best channel.
func Score(req models.DispositionRequest) models.DispositionResponse {
	grade := req.Passport.GradeNumeric
	reasons := []string{}
	guardrails := []string{}

	deliveryKm := defaultDelivery
	demandScore := 0.0
	wishCount := 0
	if req.Demand != nil {
		demandScore = req.Demand.DemandScore
		wishCount = req.Demand.OpenWishCount
		if req.Demand.NearestKm > 0 {
			deliveryKm = req.Demand.NearestKm
		}
	}

	// Guardrail: chain-depth cap is a hard override.
	if req.TransferCount >= chainDepthCap {
		guardrails = append(guardrails, fmt.Sprintf("chain_depth_cap(%d>=%d)", req.TransferCount, chainDepthCap))
		channel := "refurb"
		switch {
		case grade < gradeFairFloor:
			channel = "recycle"
		case grade < gradeGoodFloor:
			channel = "donate"
		}
		reasons = append(reasons, "transfer cap reached → end-of-cycle channel")
		return finalize(channel, grade, 0, reasons, guardrails, deliveryKm)
	}

	// Exchange-first: size/fit reasons skip the warehouse when stock exists.
	if sizeReasons[req.ReturnReason] {
		if req.ExchangeAvailable {
			reasons = append(reasons, "size/fit reason + exchange SKU in stock → exchange-first")
			return finalize("exchange", grade, demandScore, reasons, guardrails, deliveryKm)
		}
		reasons = append(reasons, "size/fit reason but no exchange stock → fall through")
	}

	// Grade-based base routing, demand re-weights resale channels.
	var channel string
	switch {
	case grade >= gradeGoodFloor:
		netRescue := carbon.NetSaved("rescue", deliveryKm)
		if demandScore > 0 && netRescue > 0 {
			channel = "rescue"
			reasons = append(reasons, fmt.Sprintf("good grade + local demand (%d wishes, score %.2f)", wishCount, demandScore))
		} else {
			channel = "p2p_resale"
			if netRescue <= 0 {
				guardrails = append(guardrails, "net_carbon_gate(rescue<=0)")
				reasons = append(reasons, "good grade but rescue net-carbon ≤ 0 → p2p")
			} else {
				reasons = append(reasons, "good grade, no local demand → p2p")
			}
		}
	case grade >= gradeFairFloor:
		channel = "refurb"
		reasons = append(reasons, "fair grade → refurb")
	default:
		channel = "donate"
		reasons = append(reasons, "low grade → donate")
	}

	return finalize(channel, grade, demandScore, reasons, guardrails, deliveryKm)
}

func finalize(channel string, grade, demandScore float64, reasons, guardrails []string, deliveryKm float64) models.DispositionResponse {
	norm := demandScore / (demandScore + 1.0) // 0..1
	score := grade*(1-demandWeight) + norm*demandWeight
	if score > 1 {
		score = 1
	}
	net := carbon.NetSaved(channel, deliveryKm)
	return models.DispositionResponse{
		Channel:           channel,
		Score:             round3(score),
		Reasons:           reasons,
		GuardrailsApplied: guardrails,
		NetCO2SavedKg:     &net,
	}
}

func round3(v float64) float64 {
	return float64(int64(v*1000+0.5)) / 1000
}
