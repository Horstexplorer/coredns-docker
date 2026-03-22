package coredns_docker

import (
	"strings"
	"sync"

	"github.com/miekg/dns"
)

type DockerDNSRecordResolver struct {
	service  *DockerService
	parser   *DNSLabelParser
	defaults *DNSLabelGroup
	cacheMtx sync.RWMutex
	cache    map[string][]dns.RR
}

func NewDockerDNSRecordResolver(service *DockerService, parser *DNSLabelParser, defaults *DNSLabelGroup) (*DockerDNSRecordResolver, error) {
	instance := &DockerDNSRecordResolver{
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

func (instance *DockerDNSRecordResolver) FindRecordsByHostname(hostname string) ([]dns.RR, bool) {
	instance.cacheMtx.RLock()
	defer instance.cacheMtx.RUnlock()

	records, ok := instance.cache[strings.ToLower(hostname)]
	return records, ok
}

func (instance *DockerDNSRecordResolver) RefreshCache() error {

	containers, err := instance.service.ListAllContainers()
	if err != nil {
		return err
	}

	temp := make(map[string][]dns.RR)
	for _, details := range containers {
		labels := details.Labels
		groups, err := ParseDNSLabelGroups(labels, instance.parser, *instance.defaults)
		if err != nil {
			return err
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

	instance.cacheMtx.Lock()
	defer instance.cacheMtx.Unlock()

	instance.cache = temp

	return nil
}
