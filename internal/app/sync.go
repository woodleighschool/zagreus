package app

import (
	"context"
	"fmt"
	"strconv"

	"github.com/woodleighschool/zagreus/external/nessus"
)

type Severity string

const (
	InfoSeverity     Severity = `info`
	LowSeverity      Severity = `low`
	MediumSeverity   Severity = `medium`
	HighSeverity     Severity = `high`
	CriticalSeverity Severity = `critical`
)

type Vulnerability struct {
	Title       string
	Description string
	Label       Severity
	Hosts       []string
}

func (s *Service) Sync(ctx context.Context) ([]Result, error) {
	hosts, err := s.nessus.GetScanDetails(ctx, s.config.Nessus.Settings.Scan)
	if err != nil {
		return nil, fmt.Errorf("unable to get nessus scans: %w", err)
	}
	var vulnerabilities map[string]Vulnerability
	for _, host := range hosts {
		for _, vuln := range host.Vulnerabilities {
			value, ok := vulnerabilities[vuln.PluginName]
			if ok {
				value.Hosts = append(value.Hosts, host.Info.FQDN)
			} else {
				payload, err := s.createPayload(ctx, vuln)
				if err != nil {

				}
			}
			vulnerabilities[vuln.PluginName] = value
		}
	}
}

func (s *Service) createPayload(ctx context.Context, vuln nessus.HostVulnerability) (Vulnerability, error) {
	var result Vulnerability
	vulnDetails, err := s.nessus.GetPluginDetails(ctx, vuln.PluginID)
	if err != nil {
		return Vulnerability{}, err
	}
	result.Title = vulnDetails.Name

	cvssScore := "unknown"
	cvssVector := "unknown"
	vulnAge := "unknown"

	if vulnDetails.CVSS3BaseScore != nil {
		cvssScore = *vulnDetails.CVSS3BaseScore
		cveBaseScore, err := strconv.ParseFloat(*vulnDetails.CVSS3BaseScore, 64)
		if err != nil {
			return Vulnerability{}, err
		}
		switch cveBaseScore {
		case 0.0 - 3.9:
			result.Label = LowSeverity
		case 4.0 - 6.9:
			result.Label = MediumSeverity
		case 7.0 - 8.9:
			result.Label = HighSeverity
		case 9.0 - 10.0:
			result.Label = CriticalSeverity
		default:
			result.Label = InfoSeverity
		}
	} else {
		result.Label = InfoSeverity
	}

	if vulnDetails.CVSS3Vector != nil {
		cvssVector = *vulnDetails.CVSS3Vector
	}

	if vulnDetails.VulnAge != nil {
		vulnAge = *vulnDetails.VulnAge
	}

	description := fmt.Sprintf(`
	%s
	CVE Score: %s
	CVE Vector: %s
	Age: %s

	External:
	`, vulnDetails.Name, cvssScore, cvssVector, vulnAge)
	for _, resource := range vulnDetails.ExternalResources {
		description = description + fmt.Sprintf("\n[%s] %s", resource)
	}

	result.Description = description
	result.Hosts = append(result.Hosts, vuln.Hostname)

	return result, nil
}
