package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/shikherjha/relay-engine/internal/config"
	"github.com/shikherjha/relay-engine/internal/handlers"
)

func main() {
	cfg := config.Load()
	h := handlers.New(cfg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/health", h.Health)

	r.Route("/disposition", func(r chi.Router) {
		r.Post("/score", h.DispositionScore())
	})
	r.Route("/match", func(r chi.Router) {
		r.Post("/rescue", h.MatchRescue())
		r.Post("/wishlist", h.MatchWishlist())
		r.Post("/pair-rescue", h.MatchPairRescue())
	})

	addr := ":" + cfg.Port
	log.Printf("relay-engine listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
