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
}

func Load() Config {
	return Config{
		Port:                getenv("PORT", "8002"),
		DatabaseURL:         getenv("DATABASE_URL", ""),
		RescueDefaultRadius: getenvFloat("RESCUE_DEFAULT_RADIUS_KM", 3.0),
		RescueDiscountBase:  getenvFloat("RESCUE_DISCOUNT_BASE", 0.15),
		RescueDiscountMax:   getenvFloat("RESCUE_DISCOUNT_MAX", 0.45),
		PairRescueRadiusKm:  getenvFloat("PAIR_RESCUE_RADIUS_KM", 10.0),
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
