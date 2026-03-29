package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

const PluginName = "coredns-docker"

var Logger = log.NewWithPlugin(PluginName)

type DNSConfiguration struct {
	Hostname  string
	AType     string
	AAAAType  string
	CNameType string
}

type Plugin struct {
	next     plugin.Handler
	config   *PluginConfig
	resolver *DNSLabelResolver
}

func NewPlugin(cfgMap map[string][]string) (*Plugin, error) {

	cfg, err := ParsePluginConfig(cfgMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin config: %w", err)
	}
	Logger.Infof("Resolved configuration:\n%s", cfg.info())

	docker, err := NewDockerService()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker service: %w", err)
	}

	parser, err := NewLabelParser(cfg.LabelPrefix, ".", cfg.UseLabelGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to create label parser: %w", err)
	}

	resolver, err := NewDNSLabelResolver(docker, parser, NewDNSRecordConverter(cfg.DNSResponseDefaults))
	if err != nil {
		return nil, fmt.Errorf("failed to create dns label resolver: %w", err)
	}

	return &Plugin{
		config:   cfg,
		resolver: resolver,
	}, nil
}

func (r *Plugin) loop(ctx context.Context) {

	Logger.Debugf("Starting background ticker with an interval of %s", r.config.UpdateInterval)

	ticker := time.NewTicker(r.config.UpdateInterval)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			Logger.Debug("Ticker ticked, running background tasks...")

			err := r.resolver.RefreshCache()
			if err != nil {
				Logger.Errorf("An error occurred: %s", err)
			}

			Logger.Debug("Background tasks completed")
		}
	}
}

func (r *Plugin) Register(next plugin.Handler) plugin.Handler {
	r.next = next
	go r.loop(context.Background())
	return r
}

func (_ *Plugin) Name() string {
	return PluginName
}

func (r *Plugin) ServeDNS(_ context.Context, writer dns.ResponseWriter, request *dns.Msg) (int, error) {

	if len(request.Question) != 1 {
		return dns.RcodeNotImplemented, nil
	}

	question := request.Question[0]

	records, err := r.resolver.FindRecordsByFQDN(question.Name)
	if err != nil || len(records) == 0 {
		return dns.RcodeNameError, nil
	}

	response := NewDNSResponse(request, true, records)

	if writer.WriteMsg(response) != nil {
		return dns.RcodeServerFailure, err
	}
	return dns.RcodeSuccess, nil
}
