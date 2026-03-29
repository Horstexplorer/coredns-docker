package docker

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/miekg/dns"
)

type DNSRecordConverterDefaults struct {
	TTL        uint32
	AValue     net.IP
	AAAAValue  net.IP
	CNAMEValue string
}

func DefaultDNSRecordConverterDefaults() *DNSRecordConverterDefaults {
	return &DNSRecordConverterDefaults{
		TTL: 0,
	}
}

type DNSRecordConverter struct {
	defaults *DNSRecordConverterDefaults
}

func NewDNSRecordConverter(defaults *DNSRecordConverterDefaults) *DNSRecordConverter {
	return &DNSRecordConverter{
		defaults: defaults,
	}
}

func (r *DNSRecordConverter) Convert(labelData map[string]string) (string, []dns.RR, error) {
	hostname, err := r.extractHostname(labelData)
	if err != nil {
		return hostname, nil, fmt.Errorf("failed to extract hostname: %w", err)
	}
	ttl, err := r.extractTTL(labelData)
	if err != nil {
		return hostname, nil, fmt.Errorf("failed to extract TTL: %w", err)
	}

	records := make([]dns.RR, 0)

	aValue, err := r.extractAValue(labelData)
	if err == nil && aValue != nil {
		aRecord, err := NewARecord(hostname, ttl, aValue)
		if err != nil {
			return hostname, records, fmt.Errorf("failed to build A record: %w", err)
		}
		records = append(records, aRecord)
	}

	aaaaValue, err := r.extractAAAAValue(labelData)
	if err == nil && aaaaValue != nil {
		aaaaRecord, err := NewAAAARecord(hostname, ttl, aaaaValue)
		if err != nil {
			return hostname, records, fmt.Errorf("failed to build AAAA record: %w", err)
		}
		records = append(records, aaaaRecord)
	}

	cnameValue, err := r.extractCNAMEValue(labelData)
	if err == nil && cnameValue != "" {
		cnameRecord, err := NewCNAMERecord(hostname, ttl, cnameValue)
		if err != nil {
			return hostname, records, fmt.Errorf("failed to build CNAME record: %w", err)
		}
		records = append(records, cnameRecord)
	}

	return hostname, records, nil
}

func (r *DNSRecordConverter) extractHostname(labelData map[string]string) (string, error) {
	hostname, ok := labelData["hostname"]
	if !ok || hostname == "" {
		return "", errors.New("missing 'hostname' in label data")
	}
	fqdn, err := ConvertToFqdn(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to convert hostname %q to FQDN: %w", hostname, err)
	}
	return fqdn, nil
}

func (r *DNSRecordConverter) extractTTL(labelData map[string]string) (uint32, error) {
	ttlString, ok := labelData["ttl"]
	if !ok {
		return r.defaults.TTL, nil
	}
	parsedTTL, err := strconv.ParseUint(ttlString, 10, 32)
	if err != nil || parsedTTL < 0 {
		return r.defaults.TTL, fmt.Errorf("invalid TTL value %q: %w", ttlString, err)
	}
	return uint32(parsedTTL), nil
}

func (r *DNSRecordConverter) extractAValue(labelData map[string]string) (net.IP, error) {
	aString, ok := labelData["a"]
	if ok && aString != "" {
		ip := net.ParseIP(aString)
		if ip == nil {
			return nil, fmt.Errorf("failed to parse IP address %q", aString)
		}
		return ip, nil
	}

	if r.defaults == nil || r.defaults.AValue == nil {
		return nil, errors.New("missing 'a' in label data and no default A value provided")
	}

	return r.defaults.AValue, nil
}

func (r *DNSRecordConverter) extractAAAAValue(labelData map[string]string) (net.IP, error) {
	aaaaString, ok := labelData["aaaa"]
	if ok && aaaaString != "" {
		ip := net.ParseIP(aaaaString)
		if ip == nil {
			return nil, fmt.Errorf("failed to parse IP address %q", aaaaString)
		}
		return ip, nil
	}

	if r.defaults == nil || r.defaults.AAAAValue == nil {
		return nil, errors.New("missing 'aaaa' in label data and no default AAAA value provided")
	}

	return r.defaults.AAAAValue, nil
}

func (r *DNSRecordConverter) extractCNAMEValue(labelData map[string]string) (string, error) {
	cnameString, ok := labelData["cname"]
	if ok && cnameString != "" {
		return cnameString, nil
	}

	if r.defaults == nil || r.defaults.CNAMEValue == "" {
		return "", errors.New("missing 'cname' in label data and no default CNAME value provided")
	}
	return r.defaults.CNAMEValue, nil
}
