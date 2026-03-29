package docker

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
)

type DNSLabelResolver struct {
	service         *DockerApiUtil
	labelParser     *LabelParser
	recordConverter *DNSRecordConverter

	cacheMtx sync.RWMutex
	cache    map[string][]dns.RR
}

func NewDNSLabelResolver(service *DockerApiUtil, labelParser *LabelParser, recordConverter *DNSRecordConverter) (*DNSLabelResolver, error) {
	Logger.Debug("Initializing DNS label record resolver...")

	instance := &DNSLabelResolver{
		service:         service,
		labelParser:     labelParser,
		recordConverter: recordConverter,
		cache:           make(map[string][]dns.RR),
	}

	if err := instance.RefreshCache(); err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *DNSLabelResolver) FindRecordsByFQDN(name string) ([]dns.RR, error) {
	r.cacheMtx.RLock()
	defer r.cacheMtx.RUnlock()

	fqdn, err := ConvertToFqdn(name)
	if err != nil {
		return nil, fmt.Errorf("failed to convert name %q to FQDN: %w", name, err)
	}

	records, ok := r.cache[fqdn]
	if !ok {
		return nil, fmt.Errorf("no DNS records found for FQDN %q", fqdn)
	}
	return records, nil
}

func (r *DNSLabelResolver) RefreshCache() error {

	Logger.Debug("Collecting container labels...")

	containers, err := r.service.ListAllContainers()
	if err != nil {
		return fmt.Errorf("failed to list docker containers: %w", err)
	}

	temp := make(map[string][]dns.RR)
	for _, details := range containers {

		for _, labelGroupData := range r.labelParser.ParseLabelGroups(details.Labels) {

			fqdn, extractedRecords, err := r.recordConverter.Convert(labelGroupData)
			if err != nil {
				Logger.Warningf("Failed to convert label group data to DNS records for container %q: %v", details.ID, err)
				continue
			}

			if _, exists := temp[fqdn]; exists {
				temp[fqdn] = append(temp[fqdn], extractedRecords...)
			} else {
				temp[fqdn] = extractedRecords
			}

			Logger.Debugf("Extracted %d DNS records for container %q with FQDN %q", len(extractedRecords), details.ID, fqdn)

		}

	}

	Logger.Debugf("Will refresh DNS record cache with %d entries", len(temp))

	r.cacheMtx.Lock()
	defer r.cacheMtx.Unlock()

	r.cache = temp

	return nil
}
