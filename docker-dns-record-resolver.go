package docker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

type DNSRecordResolver struct {
	service  *DockerApiUtil
	parser   *DNSLabelParser
	defaults *DNSLabelGroup
	cacheMtx sync.RWMutex
	cache    map[string][]dns.RR
}

func NewDockerDNSRecordResolver(service *DockerApiUtil, parser *DNSLabelParser, defaults *DNSLabelGroup) (*DNSRecordResolver, error) {
	instance := &DNSRecordResolver{
		service:  service,
		parser:   parser,
		defaults: defaults,
		cache:    make(map[string][]dns.RR),
	}

	if err := instance.RefreshCache(); err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *DNSRecordResolver) FindRecordsByHostname(hostname string) ([]dns.RR, bool) {
	r.cacheMtx.RLock()
	defer r.cacheMtx.RUnlock()

	records, ok := r.cache[strings.ToLower(hostname)]
	return records, ok
}

func (r *DNSRecordResolver) RefreshCache() error {

	containers, err := r.service.ListAllContainers()
	if err != nil {
		return fmt.Errorf("failed to list docker containers: %w", err)
	}

	temp := make(map[string][]dns.RR)
	for _, details := range containers {
		labels := details.Labels
		groups, err := ParseDNSLabelGroups(labels, r.parser, *r.defaults)
		if err != nil {
			return fmt.Errorf("failed to parse dns label groups for container %q: %w", details.ID, err)
		}

		extractedRecords := make(map[string][]dns.RR)
		for key, entry := range groups {
			records, err := entry.toDNSRecords()
			if err != nil {
				continue
			}
			extractedRecords[key] = append(extractedRecords[key], records...)
		}

		temp = Merge(temp, extractedRecords)
	}

	r.cacheMtx.Lock()
	defer r.cacheMtx.Unlock()

	r.cache = temp

	return nil
}
