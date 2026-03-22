package coredns_docker

import (
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

func (instance *DNSLabelGroup) isValid() error {
	if instance.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if instance.AValue == "" && instance.AAAAValue == "" && instance.CNAMEValue == "" {
		return fmt.Errorf("a value, aaaa value, or cname value is required")
	}
	if instance.TTL < 0 {
		return fmt.Errorf("ttl cannot be negative")
	}
	return nil
}

func (instance *DNSLabelGroup) toDNSRecords() ([]dns.RR, error) {
	var records []dns.RR

	if err := instance.isValid(); err != nil {
		return nil, err
	}

	if instance.AValue != "" {
		records = append(records, ARecord(instance.Hostname, instance.TTL, instance.AValue))
	}
	if instance.AAAAValue != "" {
		records = append(records, AAAARecord(instance.Hostname, instance.TTL, instance.AAAAValue))
	}
	if instance.CNAMEValue != "" {
		records = append(records, CNAMERecord(instance.Hostname, instance.TTL, instance.CNAMEValue))
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

	for _, key := range Keys(labels) {
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
