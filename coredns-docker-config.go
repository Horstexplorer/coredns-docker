package docker

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type CoreDNSDockerConfig struct {
	UpdateInterval time.Duration
	Defaults       *DNSLabelGroup
}

func DefaultCoreDNSDockerConfig() *CoreDNSDockerConfig {
	return &CoreDNSDockerConfig{
		UpdateInterval: 30 * time.Second,
		Defaults: &DNSLabelGroup{
			TTL: 0,
		},
	}
}

func NewCoreDNSDockerConfig(raw map[string][]string) (*CoreDNSDockerConfig, error) {

	config := DefaultCoreDNSDockerConfig()

	// update interval
	rawUpdateInterval, ok := raw["update-interval"]
	if !ok || len(rawUpdateInterval) != 1 {
		return nil, errors.New("update-interval must be provided and contain exactly one value")
	}
	parsedDuration, err := time.ParseDuration(rawUpdateInterval[0])
	if err != nil {
		return nil, fmt.Errorf("invalid update-interval %q: %w", rawUpdateInterval[0], err)
	}
	config.UpdateInterval = parsedDuration

	// ttl
	rawDefaultTTL, ok := raw["default-ttl"]
	if ok {
		if len(rawDefaultTTL) != 1 {
			return nil, errors.New("default-ttl must contain exactly one value if provided")
		}
		parsedTTL, err := strconv.ParseUint(rawDefaultTTL[0], 10, 32)
		if err != nil || parsedTTL < 0 {
			return nil, fmt.Errorf("invalid default-ttl %q: %w", rawDefaultTTL[0], err)
		}
		config.Defaults.TTL = uint32(parsedTTL)
	}

	// records
	rawDefaultCName, ok := raw["default-cname"]
	if ok {
		if len(rawDefaultTTL) != 1 {
			return nil, errors.New("default-cname must contain exactly one value if provided")
		}
		config.Defaults.CNAMEValue = rawDefaultCName[0]
	}
	rawDefaultA, ok := raw["default-a"]
	if ok {
		if len(rawDefaultA) != 1 {
			return nil, errors.New("default-a must contain exactly one value if provided")
		}
		config.Defaults.AValue = rawDefaultA[0]
	}
	rawDefaultAAAA, ok := raw["default-aaaa"]
	if ok {
		if len(rawDefaultAAAA) != 1 {
			return nil, errors.New("default-aaaa must contain exactly one value if provided")
		}
		config.Defaults.AAAAValue = rawDefaultAAAA[0]
	}

	return config, nil
}
