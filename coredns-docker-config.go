package coredns_docker

import "time"

type CoreDNSDockerConfig struct {
	UpdateInterval time.Duration
	Parser         DNSLabelParser
	Defaults       DNSLabelGroup
}

func NewCoreDNSDockerConfig(raw map[string][]string) (*CoreDNSDockerConfig, error) {
	// todo: parse config
	return &CoreDNSDockerConfig{}, nil
}
