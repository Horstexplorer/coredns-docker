package docker

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type PluginConfig struct {
	UpdateInterval      time.Duration
	LabelPrefix         string
	UseLabelGroups      bool
	DNSResponseDefaults *DNSRecordConverterDefaults
}

func (r *PluginConfig) info() string {
	return fmt.Sprintf(`
		- Update interval: %s
		- Label Prefix: %s
		- Use Label Groups: %t
		- DNSResponseDefaults
			- CNAME value: %q
			- A value: %q
			- AAAA value: %q
			- TTL: %d`,
		r.UpdateInterval,
		r.LabelPrefix,
		r.UseLabelGroups,
		r.DNSResponseDefaults.CNAMEValue,
		r.DNSResponseDefaults.AValue,
		r.DNSResponseDefaults.AAAAValue,
		r.DNSResponseDefaults.TTL)
}

func DefaultPluginConfig() *PluginConfig {
	return &PluginConfig{
		UpdateInterval:      30 * time.Second,
		LabelPrefix:         PluginName,
		UseLabelGroups:      true,
		DNSResponseDefaults: DefaultDNSRecordConverterDefaults(),
	}
}

func ParsePluginConfig(raw map[string][]string) (*PluginConfig, error) {

	Logger.Debug("Parsing plugin configuration...")

	config := DefaultPluginConfig()

	// update interval
	rawUpdateInterval, ok := raw["update-interval"]
	if ok {
		if len(rawUpdateInterval) != 1 {
			return nil, errors.New("update-interval must contain exactly one value")
		}
		parsedDuration, err := time.ParseDuration(rawUpdateInterval[0])
		if err != nil {
			return nil, fmt.Errorf("invalid update-interval %q: %w", rawUpdateInterval[0], err)
		}
		config.UpdateInterval = parsedDuration
	}

	// label prefix
	rawLabelPrefix, ok := raw["label-prefix"]
	if ok {
		if len(rawLabelPrefix) != 1 {
			return nil, errors.New("label-prefix must contain exactly one value")
		}
		if rawLabelPrefix[0] == "" {
			return nil, fmt.Errorf("invalid label-prefix: %q", rawLabelPrefix[0])
		}
		config.LabelPrefix = rawLabelPrefix[0]
	}

	// use label groups
	rawUseLabelGroups, ok := raw["use-label-groups"]
	if ok {
		if len(rawUseLabelGroups) != 1 {
			return nil, errors.New("update-interval must contain exactly one value")
		}
		parsedUseLabelGroups, err := strconv.ParseBool(rawUseLabelGroups[0])
		if err != nil {
			return nil, fmt.Errorf("invalid use-label-groups %q: %w", rawUseLabelGroups[0], err)
		}
		config.UseLabelGroups = parsedUseLabelGroups
	}

	// records
	rawDefaultCName, ok := raw["default-cname"]
	if ok {
		if len(rawDefaultCName) != 1 {
			return nil, errors.New("default-cname must contain exactly one value if provided")
		}
		config.DNSResponseDefaults.CNAMEValue = rawDefaultCName[0]
	}

	rawDefaultA, ok := raw["default-a"]
	if ok {
		if len(rawDefaultA) != 1 {
			return nil, errors.New("default-a must contain exactly one value if provided")
		}
		ipv4 := net.ParseIP(rawDefaultA[0])
		if ipv4 == nil || len(ipv4) != net.IPv4len {
			return nil, fmt.Errorf("invalid default-a IP address %q", rawDefaultA[0])
		}
		config.DNSResponseDefaults.AValue = ipv4
	}

	rawDefaultAAAA, ok := raw["default-aaaa"]
	if ok {
		if len(rawDefaultAAAA) != 1 {
			return nil, errors.New("default-aaaa must contain exactly one value if provided")
		}
		ipv6 := net.ParseIP(rawDefaultAAAA[0])
		if ipv6 == nil || len(ipv6) != net.IPv6len {
			return nil, fmt.Errorf("invalid default-aaaa IP address %q", rawDefaultAAAA[0])
		}
		config.DNSResponseDefaults.AAAAValue = ipv6
	}

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
		config.DNSResponseDefaults.TTL = uint32(parsedTTL)
	}

	return config, nil
}
