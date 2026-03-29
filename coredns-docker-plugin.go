package docker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

const PluginName = "coredns-docker"

type DNSConfiguration struct {
	Hostname  string
	AType     string
	AAAAType  string
	CNameType string
}

type CoreDNSDockerPlugin struct {
	next     plugin.Handler
	config   *CoreDNSDockerConfig
	resolver *DNSRecordResolver
}

func NewCoreDNSDockerPlugin(cfgMap map[string][]string) (*CoreDNSDockerPlugin, error) {
	cfg, err := NewCoreDNSDockerConfig(cfgMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin config: %w", err)
	}

	docker, err := NewDockerService()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker service: %w", err)
	}

	dnsRecordResolver, err := NewDockerDNSRecordResolver(docker, &DNSLabelParser{
		pattern:   *regexp.MustCompile(`^coredns\.(?P<id>.+)\.(?P<type>.+)$`),
		idGroup:   "id",
		typeGroup: "type",
	}, cfg.Defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to create dns record resolver: %w", err)
	}

	return &CoreDNSDockerPlugin{
		config:   cfg,
		resolver: dnsRecordResolver,
	}, nil
}

func (r *CoreDNSDockerPlugin) loop(ctx context.Context) {
	ticker := time.NewTicker(r.config.UpdateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := r.resolver.RefreshCache()
			if err != nil {
				log.Printf("An error occurred: %s", err)
			}
		}
	}
}

func (r *CoreDNSDockerPlugin) Register(next plugin.Handler) plugin.Handler {
	r.next = next
	go r.loop(context.Background())
	return r
}

func (_ *CoreDNSDockerPlugin) Name() string {
	return PluginName
}

func (r *CoreDNSDockerPlugin) ServeDNS(ctx context.Context, writer dns.ResponseWriter, request *dns.Msg) (int, error) {
	if len(request.Question) == 0 {
		return plugin.NextOrFailure(PluginName, r.next, ctx, writer, request)
	} else if len(request.Question) > 1 {
		return dns.RcodeNotImplemented, nil
	}

	question := request.Question[0]

	records, ok := r.resolver.FindRecordsByHostname(question.Name)
	if !ok {
		return plugin.NextOrFailure(PluginName, r.next, ctx, writer, request)
	}

	matchingRecords := make([]dns.RR, 0)
	for _, record := range records {
		if record.Header().Rrtype == question.Qtype {
			matchingRecords = append(matchingRecords, record)
		}
	}

	if len(matchingRecords) <= 0 {
		return plugin.NextOrFailure(PluginName, r.next, ctx, writer, request)
	}

	response := BuildDNSResponse(request, true, matchingRecords)

	if err := writer.WriteMsg(response); err != nil {
		return dns.RcodeServerFailure, err
	}
	return dns.RcodeSuccess, nil
}
