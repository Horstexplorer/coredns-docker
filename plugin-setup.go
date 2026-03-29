package docker

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	plugin.Register(PluginName, setup)
}

func setup(c *caddy.Controller) error {

	Logger.Debug("Starting plugin setup...")

	rawCfg, err := parseConfig(c)
	if err != nil {
		return plugin.Error(PluginName, err)
	}

	coreDNSDockerPlugin, err := NewPlugin(rawCfg)
	if err != nil {
		return plugin.Error(PluginName, err)
	}

	Logger.Debug("Registering plugin with CoreDNS...")

	dnsserver.GetConfig(c).AddPlugin(coreDNSDockerPlugin.Register)

	return nil
}

func parseConfig(c *caddy.Controller) (map[string][]string, error) {
	var configMap = make(map[string][]string)
	for c.Next() {
		for c.NextBlock() {
			configMap[c.Val()] = c.RemainingArgs()
		}
	}
	return configMap, nil
}
