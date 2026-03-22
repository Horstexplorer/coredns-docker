package coredns_docker

import (
	"context"

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
	next   plugin.Handler
	config *CoreDNSDockerConfig

	resolver *DockerDNSRecordResolver
}

func NewCoreDNSDockerPlugin(cfgMap map[string][]string) (*CoreDNSDockerPlugin, error) {
	cfg, err := NewCoreDNSDockerConfig(cfgMap)
	if err != nil {
		return nil, err
	}

	docker, err := NewDockerService()
	if err != nil {
		return nil, err
	}

	dnsRecordResolver, err := NewDockerDNSRecordResolver(docker, &cfg.Parser, &cfg.Defaults)
	if err != nil {
		return nil, err
	}

	return &CoreDNSDockerPlugin{
		config:   cfg,
		resolver: dnsRecordResolver,
	}, nil
}

func (instance *CoreDNSDockerPlugin) Register(next plugin.Handler) plugin.Handler {
	instance.next = next
	// todo: should start background goroutine to refresh cache periodically
	return instance
}

func (_ *CoreDNSDockerPlugin) Name() string {
	return PluginName
}

func (instance *CoreDNSDockerPlugin) ServeDNS(ctx context.Context, writer dns.ResponseWriter, request *dns.Msg) (int, error) {
	if len(request.Question) == 0 {
		return plugin.NextOrFailure(PluginName, instance.next, ctx, writer, request)
	}

	question := request.Question[0]

	records, ok := instance.resolver.FindRecordsByHostname(question.Name)
	if !ok {
		return plugin.NextOrFailure(PluginName, instance.next, ctx, writer, request)
	}

	matchingRecords := make([]dns.RR, 0)
	for _, record := range records {
		if record.Header().Rrtype == question.Qtype {
			matchingRecords = append(matchingRecords, record)
		}
	}

	if len(matchingRecords) <= 0 {
		return plugin.NextOrFailure(PluginName, instance.next, ctx, writer, request)
	}

	response := BuildDNSResponse(request, true, matchingRecords)

	if err := writer.WriteMsg(response); err != nil {
		return dns.RcodeServerFailure, err
	}
	return dns.RcodeSuccess, nil
}
