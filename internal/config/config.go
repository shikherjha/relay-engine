package config

import (
	"os"
	"strconv"
)

// Config holds runtime settings sourced from the environment.
type Config struct {
	Port                string
	DatabaseURL         string
	RescueDefaultRadius float64
	RescueDiscountBase  float64
	RescueDiscountMax   float64
	PairRescueRadiusKm  float64
	// DispositionScorer selects the routing policy: "rules" (default) or "rl"
	// (future RL policy; falls back to rules until trained). See engine-rl-hook.
	DispositionScorer string

	// Rescue Dispatch Score term weights (plan.md §21.4). Positives ideally sum
	// to 1.0; the two risk terms subtract. Keep in sync with the relay-api mock.
	DispatchWDemand    float64
	DispatchWDistance  float64
	DispatchWTTL       float64
	DispatchWPrice     float64
	DispatchWKeep      float64
	DispatchWCarbon    float64
	DispatchWFailRisk  float64
	DispatchWChainRisk float64
}

func Load() Config {
	return Config{
		Port:                getenv("PORT", "8002"),
		DatabaseURL:         getenv("DATABASE_URL", ""),
		RescueDefaultRadius: getenvFloat("RESCUE_DEFAULT_RADIUS_KM", 3.0),
		RescueDiscountBase:  getenvFloat("RESCUE_DISCOUNT_BASE", 0.15),
		RescueDiscountMax:   getenvFloat("RESCUE_DISCOUNT_MAX", 0.45),
		PairRescueRadiusKm:  getenvFloat("PAIR_RESCUE_RADIUS_KM", 10.0),
		DispositionScorer:   getenv("DISPOSITION_SCORER", "rules"),
		DispatchWDemand:     getenvFloat("DISPATCH_W_DEMAND", 0.28),
		DispatchWDistance:   getenvFloat("DISPATCH_W_DISTANCE", 0.18),
		DispatchWTTL:        getenvFloat("DISPATCH_W_TTL", 0.12),
		DispatchWPrice:      getenvFloat("DISPATCH_W_PRICE", 0.12),
		DispatchWKeep:       getenvFloat("DISPATCH_W_KEEP", 0.15),
		DispatchWCarbon:     getenvFloat("DISPATCH_W_CARBON", 0.15),
		DispatchWFailRisk:   getenvFloat("DISPATCH_W_FAIL_RISK", 0.5),
		DispatchWChainRisk:  getenvFloat("DISPATCH_W_CHAIN_RISK", 0.2),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
