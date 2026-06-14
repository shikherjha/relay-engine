package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shikherjha/relay-engine/internal/config"
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

// Health reports liveness for compose/healthcheck.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "relay-engine",
	})
}

// notImplemented returns a stub for routes whose logic lands in T1/T2.
// Endpoint signatures mirror relay-contracts (plan.md §5).
func notImplemented(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotImplemented, map[string]any{
			"error":    "not_implemented",
			"endpoint": name,
			"note":     "T0 skeleton — logic implemented in T1/T2.",
		})
	}
}

// DispositionScore — POST /disposition/score  (T1: engine-disposition)
func (h *Handler) DispositionScore() http.HandlerFunc { return notImplemented("/disposition/score") }

// MatchRescue — POST /match/rescue  (T1: engine-rescue-ttl)
func (h *Handler) MatchRescue() http.HandlerFunc { return notImplemented("/match/rescue") }

// MatchWishlist — POST /match/wishlist  (T1: engine-match-vector)
func (h *Handler) MatchWishlist() http.HandlerFunc { return notImplemented("/match/wishlist") }

// MatchPairRescue — POST /match/pair-rescue  (T2: engine-pair-rescue)
func (h *Handler) MatchPairRescue() http.HandlerFunc { return notImplemented("/match/pair-rescue") }
