package nessus

// Enums

type Severity int

const (
	InfoSeverity     Severity = 0
	LowSeverity      Severity = 1
	MediumSeverity   Severity = 2
	HighSeverity     Severity = 3
	CriticalSeverity Severity = 4
)

func (s Severity) String() string {
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
		ACLs            []Permission `json:"acls"`
		EditAllowed     bool         `json:"edit_allowed"`
		Status          string       `json:"status"`
		Policy          string       `json:"policy"`
		PCICanUpload    bool         `json:"pci-can-upload"`
		HasAuditTrail   bool         `json:"hasaudittrail"`
		ScanStart       int          `json:"scan_start"`
		FolderID        int          `json:"folder_id"`
		Targets         string       `json:"targets"`
		Timestamp       int          `json:"timestamp"`
		ObjectID        int          `json:"object_id"`
		ScannerName     string       `json:"scanner_name"`
		HasKB           bool         `json:"haskb"`
		UUID            string       `json:"uuid"`
		HostCount       int          `json:"hostcount"`
		ScanEnd         int          `json:"scan_end"`
		Name            string       `json:"name"`
		UserPermissions int          `json:"user_permissions"`
		Control         bool         `json:"control"`
	} `json:"info"`
	Hosts        []Host `json:"hosts"`
	CompHosts    []Host `json:"comphosts"`
	Notes        []Note `json:"notes"`
	Remediations struct {
		Remediations   []Remediation `json:"remediations"`
		Hosts          int           `json:"num_hosts"`
		CVEs           int           `json:"num_cves"`
		ImpactedHosts  int           `json:"num_impacted_hosts"`
		RemediatedCVEs int           `json:"num_remediated_cves"`
	} `json:"remediations"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Compliance      []Vulnerability `json:"compliance"`
	History         []History       `json:"history"`
	Filters         []Filter        `json:"filters"`
}

type Permission struct {
	Owner       int                  `json:"owner"`
	Type        string               `json:"type"`
	Permissions UserAccessPermission `json:"permissions"`
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
}

type UserAccessPermission int

const (
	UserBasic               UserAccessPermission = 16
	UserStandard            UserAccessPermission = 32
	UserAdministrator       UserAccessPermission = 64
	UserSystemAdministrator UserAccessPermission = 128
)

type Host struct {
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
type Note struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}
type Remediation struct {
	Value           string `json:"value"`
	Remediation     string `json:"remediation"`
	Hosts           int    `json:"hosts"`
	Vulnerabilities int    `json:"vulns"`
}
type Vulnerability struct {
	PluginID           int    `json:"plugin_id"`
	PluginName         string `json:"plugin_name"`
	PluginFamily       string `json:"plugin_family"`
	Count              int    `json:"count"`
	VulnerabilityIndex int    `json:"vuln_index"`
	SeverityIndex      int    `json:"severity_index"`
}
type History struct {
	ID               int    `json:"history_id"`
	UUID             string `json:"uuid"`
	OwnerID          int    `json:"owner_id"`
	Status           string `json:"status"`
	CreationDate     int    `json:"creation_date"`
	LastModifiedDate int    `json:"last_modification_date"`
}
type Filter struct {
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
	HostID        int      `json:"host_id"`
	Hostname      string   `json:"hostname"`
	PluginID      int      `json:"plugin_id"`
	PluginName    string   `json:"plugin_name"`
	PluginFamily  string   `json:"plugin_family"`
	Count         int      `json:"count"`
	SeverityIndex int      `json:"severity_index"`
	Severity      Severity `json:"severity"`
}
type HostVulnerability struct {
	HostID             int      `json:"host_id"`
	Hostname           string   `json:"hostname"`
	PluginID           int      `json:"plugin_id"`
	PluginName         string   `json:"plugin_name"`
	PluginFamily       string   `json:"plugin_family"`
	Count              int      `json:"count"`
	VulnerabilityIndex int      `json:"vuln_index"`
	SeverityIndex      int      `json:"severity_index"`
	Severity           Severity `json:"severity"`
}
