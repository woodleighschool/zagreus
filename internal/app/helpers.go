package app

import (
	"maps"
	"slices"

	"github.com/woodleighschool/zagreus/external/nessus"
)

func (s *Service) shouldIgnoreHost(host nessus.HostDetails) bool {
	// Check config for a host entry
	hostConfig, ok := s.config.Nessus.Settings.Ignore.Hosts[host.Info.FQDN]
	if ok {
		if hostConfig.Full != nil && *hostConfig.Full {
			return true
		}
	}
	return false
}

func (s *Service) shouldIgnoreVuln(host nessus.HostDetails, vuln nessus.HostVulnerability) bool {
	if slices.Contains(s.config.Nessus.Settings.Ignore.Severity, vuln.Severity.String()) || slices.Contains(s.config.Nessus.Settings.Ignore.Plugins, vuln.PluginID) {
		return true
	}

	hostConfig, ok := s.config.Nessus.Settings.Ignore.Hosts[host.Info.FQDN]
	if !ok {
		return false
	}
	if slices.Contains(hostConfig.Plugins, vuln.PluginID) || slices.Contains(hostConfig.Severity, vuln.Severity.String()) {
		return true
	}
	return false
}

func reconcileLists(nessusList map[string]bool, trelloList map[string]bool) (map[string]bool, bool) {
	result := make(map[string]bool)
	maps.Copy(result, trelloList)

	// Import new items from Nessus
	for host := range nessusList {
		_, ok := result[host]
		if !ok {
			result[host] = false
		}
	}

	// Mark any items in Trello not in Nessus as resolved
	for host := range trelloList {
		_, ok := nessusList[host]
		if !ok {
			result[host] = true
		}
	}

	// Check for equality with Trello
	equal := len(result) == len(trelloList)
	for host, value := range result {
		val, ok := trelloList[host]
		if !ok || val != value {
			equal = false
		}
	}

	return result, equal
}

func getIdentifier(host nessus.HostDetails) string {
	if host.Info.FQDN != "" {
		return host.Info.FQDN
	}
	return host.Info.IP
}
