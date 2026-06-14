// Package models holds the typed request/response bodies for relay-engine.
// Shapes mirror relay-contracts v1 and the relay-api Pydantic schemas so the
// HTTP boundary is identical on both sides.
package models

// ---- shared ----

type Geo struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Defect struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Bbox        []float64 `json:"bbox,omitempty"`
	Confidence  *float64 `json:"confidence,omitempty"`
	Description string   `json:"description,omitempty"`
}

// ConditionPassport is the subset the engine needs for scoring.
type ConditionPassport struct {
	UnitID          string   `json:"unit_id"`
	ReturnID        string   `json:"return_id,omitempty"`
	Grade           string   `json:"grade"`
	GradeNumeric    float64  `json:"grade_numeric"`
	Category        string   `json:"category,omitempty"`
	Vertical        string   `json:"vertical"`
	DispositionHint string   `json:"disposition_hint,omitempty"`
	Defects         []Defect `json:"defects,omitempty"`
	PackagingState  string   `json:"packaging_state,omitempty"`
	Confidence      float64  `json:"confidence"`
}

// ---- disposition ----

type DemandSignal struct {
	OpenWishCount int     `json:"open_wish_count"`
	DemandScore   float64 `json:"demand_score"`
}

type DispositionRequest struct {
	UnitID       string            `json:"unit_id"`
	Passport     ConditionPassport `json:"passport"`
	ReturnReason string            `json:"return_reason"`
	UserID       string            `json:"user_id,omitempty"`
	Geo          *Geo              `json:"geo,omitempty"`
	Demand       *DemandSignal     `json:"demand,omitempty"`
}

type DispositionResponse struct {
	Channel           string   `json:"channel"`
	Score             float64  `json:"score"`
	Reasons           []string `json:"reasons"`
	GuardrailsApplied []string `json:"guardrails_applied"`
	NetCO2SavedKg     *float64 `json:"net_co2_saved_kg,omitempty"`
}

// ---- rescue match ----

type RescueMatchRequest struct {
	UnitID   string `json:"unit_id"`
	Geo      *Geo   `json:"geo,omitempty"`
	RadiusKm float64 `json:"radius_km,omitempty"`
}

type RescueCandidate struct {
	UnitID     string  `json:"unit_id"`
	DistanceKm float64 `json:"distance_km"`
	Score      float64 `json:"score"`
}

type RescueMatchResponse struct {
	Eligible   bool              `json:"eligible"`
	Candidates []RescueCandidate `json:"candidates"`
}

// ---- wishlist match (pgvector cosine x wish_score) ----

type WishlistMatchRequest struct {
	UnitID    string    `json:"unit_id"`
	Embedding []float32 `json:"embedding"`
	Geo       *Geo      `json:"geo,omitempty"`
	RadiusKm  float64   `json:"radius_km,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

type WishMatch struct {
	WishID     string  `json:"wish_id"`
	UserID     string  `json:"user_id"`
	Score      float64 `json:"score"`
	DistanceKm float64 `json:"distance_km"`
}

type WishlistMatchResponse struct {
	Matches []WishMatch `json:"matches"`
}

// ---- pair rescue (bipartite A<->B swap) ----

type PairRescueRequest struct {
	RadiusKm float64 `json:"radius_km,omitempty"`
}

type PairMatch struct {
	UnitA string  `json:"unit_a"`
	UnitB string  `json:"unit_b"`
	UserA string  `json:"user_a"`
	UserB string  `json:"user_b"`
	Score float64 `json:"score"`
}

type PairRescueResponse struct {
	Pairs []PairMatch `json:"pairs"`
}

// Error is the shared error envelope.
type Error struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}
