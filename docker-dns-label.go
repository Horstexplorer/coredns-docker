package docker

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type DNSLabelParser struct {
	pattern   regexp.Regexp // ^(prefix)\.(?P<id>.+)\.(?P<type>.+)$
	idGroup   string        // id
	typeGroup string        // type
}

type DNSLabelGroup struct {
	Hostname   string
	AValue     string
	AAAAValue  string
	CNAMEValue string
	TTL        uint32
}

func (r *DNSLabelGroup) isValid() error {
	if r.Hostname == "" {
		return errors.New("hostname is required")
	}
	if r.AValue == "" && r.AAAAValue == "" && r.CNAMEValue == "" {
		return errors.New("a value, aaaa value, or cname value is required")
	}
	if r.TTL < 0 {
		return errors.New("ttl cannot be negative")
	}
	return nil
}

func (r *DNSLabelGroup) toDNSRecords() ([]dns.RR, error) {
	var records []dns.RR

	if err := r.isValid(); err != nil {
		return nil, err
	}

	if r.AValue != "" {
		records = append(records, ARecord(r.Hostname, r.TTL, r.AValue))
	}
	if r.AAAAValue != "" {
		records = append(records, AAAARecord(r.Hostname, r.TTL, r.AAAAValue))
	}
	if r.CNAMEValue != "" {
		records = append(records, CNAMERecord(r.Hostname, r.TTL, r.CNAMEValue))
	}

	return records, nil
}

func ParseDNSLabelGroups(labels map[string]string, parser *DNSLabelParser, defaults DNSLabelGroup) (map[string]DNSLabelGroup, error) {

	regex := parser.pattern
	idPos := regex.SubexpIndex(parser.idGroup)
	typePos := regex.SubexpIndex(parser.typeGroup)
	if idPos <= 0 || typePos <= 0 {
		return nil, fmt.Errorf("invalid regex pattern: missing required id or type group")
	}

	groups := make(map[string]DNSLabelGroup)

	for key, _ := range labels {
		if !parser.pattern.MatchString(key) {
			continue
		}
		matches := regex.FindStringSubmatch(key)

		idStr := strings.ToLower(matches[idPos])
		if _, ok := labels[idStr]; !ok {
			groups[idStr] = defaults
		}

		entry := groups[idStr]

		typeStr := strings.ToLower(matches[typePos])
		switch typeStr {
		case "hostname":
			entry.Hostname = labels[key]
		case "a":
			entry.AValue = labels[key]
		case "aaaa":
			entry.AAAAValue = labels[key]
		case "cname":
			entry.CNAMEValue = labels[key]
		case "ttl":
			value, e := strconv.ParseUint(labels[key], 10, 32)
			if e != nil {
				entry.TTL = uint32(value)
			}
		default:
			continue
		}

		groups[idStr] = entry
	}

	for key, group := range groups {
		if err := group.isValid(); err != nil {
			return nil, fmt.Errorf("invalid content for group %q: %w", key, err)
		}
	}

	return groups, nil
}
