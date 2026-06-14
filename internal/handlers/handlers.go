package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shikherjha/relay-engine/internal/config"
	"github.com/shikherjha/relay-engine/internal/models"
)

type Handler struct {
	cfg config.Config
}

func New(cfg config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, models.Error{Error: "invalid_body", Detail: err.Error()})
		return false
	}
	return true
}

// Health reports liveness for compose/healthcheck.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "relay-engine",
	})
}

// DispositionScore — POST /disposition/score  (scoring logic lands in T1).
func (h *Handler) DispositionScore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.DispositionRequest
		if !decode(w, r, &req) {
			return
		}
		resp := models.DispositionResponse{
			Channel:           "",
			Score:             0,
			Reasons:           []string{"T0 skeleton — scoring lands in T1 (engine-disposition)"},
			GuardrailsApplied: []string{},
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// MatchRescue — POST /match/rescue  (geo + TTL logic lands in T1).
func (h *Handler) MatchRescue() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RescueMatchRequest
		if !decode(w, r, &req) {
			return
		}
		writeJSON(w, http.StatusOK, models.RescueMatchResponse{
			Eligible:   false,
			Candidates: []models.RescueCandidate{},
		})
	}
}

// MatchWishlist — POST /match/wishlist  (pgvector cosine lands in T1).
func (h *Handler) MatchWishlist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.WishlistMatchRequest
		if !decode(w, r, &req) {
			return
		}
		writeJSON(w, http.StatusOK, models.WishlistMatchResponse{
			Matches: []models.WishMatch{},
		})
	}
}

// MatchPairRescue — POST /match/pair-rescue  (bipartite match lands in T2).
func (h *Handler) MatchPairRescue() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PairRescueRequest
		if !decode(w, r, &req) {
			return
		}
		writeJSON(w, http.StatusOK, models.PairRescueResponse{
			Pairs: []models.PairMatch{},
		})
	}
}
