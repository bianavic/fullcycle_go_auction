package config

import (
	"os"
	"time"
)

func ParseDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

// AuctionInterval is the duration after which an auction is automatically closed.
func AuctionInterval() time.Duration {
	return ParseDuration("AUCTION_INTERVAL", 5*time.Minute)
}
