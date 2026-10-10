package nessus

import (
	"encoding/json"
	"fmt"
)

// Enums

type ScanUserAccessPermission int

const (
	UserBasic               ScanUserAccessPermission = 16
	UserStandard            ScanUserAccessPermission = 32
	UserAdministrator       ScanUserAccessPermission = 64
	UserSystemAdministrator ScanUserAccessPermission = 128
)

type HostVulnSeverity int

const (
	InfoSeverity     HostVulnSeverity = 0
	LowSeverity      HostVulnSeverity = 1
	MediumSeverity   HostVulnSeverity = 2
	HighSeverity     HostVulnSeverity = 3
	CriticalSeverity HostVulnSeverity = 4
)

func (s HostVulnSeverity) String() string {
	switch s {
	case InfoSeverity:
		return "info"
	case LowSeverity:
		return "low"
	case MediumSeverity:
		return "medium"
	case HighSeverity:
		return "high"
	case CriticalSeverity:
		return "critical"
	default:
		return ""
	}
}

// /scans/{scan_id}

type Scan struct {
	Info struct {
		ACLs            []ScanPermission `json:"acls"`
		EditAllowed     bool             `json:"edit_allowed"`
		Status          string           `json:"status"`
		Policy          string           `json:"policy"`
		PCICanUpload    bool             `json:"pci-can-upload"`
		HasAuditTrail   bool             `json:"hasaudittrail"`
		ScanStart       int              `json:"scan_start"`
		FolderID        int              `json:"folder_id"`
		Targets         string           `json:"targets"`
		Timestamp       int              `json:"timestamp"`
		ObjectID        int              `json:"object_id"`
		ScannerName     string           `json:"scanner_name"`
		HasKB           bool             `json:"haskb"`
		UUID            string           `json:"uuid"`
		HostCount       int              `json:"hostcount"`
		ScanEnd         int              `json:"scan_end"`
		Name            string           `json:"name"`
		UserPermissions int              `json:"user_permissions"`
		Control         bool             `json:"control"`
	} `json:"info"`
	Hosts        []ScanHost `json:"hosts"`
	CompHosts    []ScanHost `json:"comphosts"`
	Notes        []ScanNote `json:"notes"`
	Remediations struct {
		Remediations   []ScanRemediation `json:"remediations"`
		Hosts          int               `json:"num_hosts"`
		CVEs           int               `json:"num_cves"`
		ImpactedHosts  int               `json:"num_impacted_hosts"`
		RemediatedCVEs int               `json:"num_remediated_cves"`
	} `json:"remediations"`
	Vulnerabilities []ScanVulnerability `json:"vulnerabilities"`
	Compliance      []ScanVulnerability `json:"compliance"`
	History         []ScanHistory       `json:"history"`
	Filters         []ScanFilter        `json:"filters"`
}

type ScanPermission struct {
	Owner       int                      `json:"owner"`
	Type        string                   `json:"type"`
	Permissions ScanUserAccessPermission `json:"permissions"`
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
}

type ScanHost struct {
	ID                  int    `json:"host_id"`
	Index               int    `json:"host_index"`
	Name                string `json:"hostname"`
	Progress            string `json:"progress"`
	Severity            int    `json:"severity"`
	Critical            int    `json:"critical"`
	High                int    `json:"high"`
	Medium              int    `json:"medium"`
	Low                 int    `json:"low"`
	Info                int    `json:"info"`
	TotalChecks         int    `json:"totalchecksconsidered"`
	ProcessedChecks     int    `json:"numchecksconsidered"`
	TotalScanProgress   int    `json:"scanprogresstotal"`
	CurrentScanProgress int    `json:"scanprogresscurrent"`
	Score               int    `json:"score"`
}
type ScanNote struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}
type ScanRemediation struct {
	Value           string `json:"value"`
	Remediation     string `json:"remediation"`
	Hosts           int    `json:"hosts"`
	Vulnerabilities int    `json:"vulns"`
}
type ScanVulnerability struct {
	PluginID           int    `json:"plugin_id"`
	PluginName         string `json:"plugin_name"`
	PluginFamily       string `json:"plugin_family"`
	Count              int    `json:"count"`
	VulnerabilityIndex int    `json:"vuln_index"`
	SeverityIndex      int    `json:"severity_index"`
}
type ScanHistory struct {
	ID               int    `json:"history_id"`
	UUID             string `json:"uuid"`
	OwnerID          int    `json:"owner_id"`
	Status           string `json:"status"`
	CreationDate     int    `json:"creation_date"`
	LastModifiedDate int    `json:"last_modification_date"`
}
type ScanFilter struct {
	ShortName string `json:"name"`
	LongName  string `json:"readable_name"`
	Operators []any  `json:"operators"`
	Control   struct {
		Type           string `json:"type"`
		ReadableRegest string `json:"readable_regest"`
		RegEx          string `json:"regex"`
		Options        []any  `json:"options"`
	} `json:"control"`
}

// /scans/{scan_id}/hosts/{host_id}

type HostDetails struct {
	Info struct {
		HostStart       string `json:"host_start"`
		MacAddress      string `json:"mac_address"`
		FQDN            string `json:"host-fqdn"`
		HostEnd         string `json:"host_end"`
		OperatingSystem string `json:"operating-system"`
		IP              string `json:"host-ip"`
	} `json:"info"`
	Compliance      []HostCompliance    `json:"compliance"`
	Vulnerabilities []HostVulnerability `json:"vulnerabilities"`
}

type HostCompliance struct {
	HostID        int              `json:"host_id"`
	Hostname      string           `json:"hostname"`
	PluginID      int              `json:"plugin_id"`
	PluginName    string           `json:"plugin_name"`
	PluginFamily  string           `json:"plugin_family"`
	Count         int              `json:"count"`
	SeverityIndex int              `json:"severity_index"`
	Severity      HostVulnSeverity `json:"severity"`
}
type HostVulnerability struct {
	HostID             int              `json:"host_id"`
	Hostname           string           `json:"hostname"`
	PluginID           int              `json:"plugin_id"`
	PluginName         string           `json:"plugin_name"`
	PluginFamily       string           `json:"plugin_family"`
	Count              int              `json:"count"`
	VulnerabilityIndex int              `json:"vuln_index"`
	SeverityIndex      int              `json:"severity_index"`
	Severity           HostVulnSeverity `json:"severity"`
}

// /plugins/plugin/{id}

type PluginDetails struct {
	Name              string
	CVSS3Vector       *string
	CVSS3BaseScore    *string
	ExternalResources []string
	VulnAge           *string
}

func (p *PluginDetails) UnmarshalJSON(b []byte) error {
	var r struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		FamilyName string `json:"family_name"`
		Attributes []struct {
			Name  string `json:"attribute_name"`
			Value string `json:"attribute_value"`
		} `json:"attributes"`
	}

	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("failed to unmarshal plugin to intermediary struct")
	}
	p.Name = r.Name
	for _, attribute := range r.Attributes {
		switch attribute.Name {
		case "cvss3_vector":
			p.CVSS3Vector = &attribute.Value
		case "cvss3_base_score":
			p.CVSS3BaseScore = &attribute.Value
		case "see_also":
			p.ExternalResources = append(p.ExternalResources, attribute.Value)
		case "age_of_vuln":
			p.VulnAge = &attribute.Value
		default:
			continue
		}
	}
	return nil
}
