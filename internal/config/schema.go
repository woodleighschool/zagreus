package config

type Config struct {
	Version  int             `yaml:"version"`
	Settings ZagreusSettings `yaml:"settings"`
	Nessus   NessusConfig    `yaml:"nessus"`
	Trello   TrelloConfig    `yaml:"trello"`
}

type ZagreusSettings struct {
	LogLevel string       `yaml:"log_level"`
	Sync     SyncSettings `yaml:"sync"`
	Terms    []Term       `yaml:"terms"`
	Mode     string       `yaml:"mode" jsonschema:"enum=create,enum=move"`
}

type SyncSettings struct {
	PollInterval Duration `yaml:"poll_interval,omitempty"`
	RetryAfter   Duration `yaml:"retry_after,omitempty"`
	RetryMax     Duration `yaml:"retry_max,omitempty"`
}

type Term struct {
	Name  string    `yaml:"name"`
	Range DateRange `yaml:"range"`
}

type NessusConfig struct {
	Connection NessusConnection `yaml:"connection"`
	Settings   NessusSettings   `yaml:"settings"`
}

type NessusConnection struct {
	Host      string `yaml:"host"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type NessusSettings struct {
	Scan   int `yaml:"scan"`
	Ignore struct {
		Severity    []string                       `yaml:"severity"`
		Plugins     []int                          `yaml:"plugins"`
		Hosts       map[string]NessusHostExclusion `yaml:"hosts"`
		Combination []string                       `yaml:"combination"`
	} `yaml:"ignore"`
}

type NessusHostExclusion struct {
	Full     *bool    `yaml:"full,omitempty"`
	Plugins  []int    `yaml:"plugins,omitempty"`
	Severity []string `yaml:"severity,omitempty"`
}

type TrelloConfig struct {
	Connection TrelloConnection `yaml:"connection"`
	Settings   TrelloSettings   `yaml:"settings"`
}

type TrelloConnection struct {
	APIKey   string `yaml:"api_key"`
	APIToken string `yaml:"api_token"`
}

type TrelloSettings struct {
	Prefix      string            `yaml:"board_prefix"`
	Lists       []string          `yaml:"lists"`
	In          string            `yaml:"in"`
	Out         string            `yaml:"out"`
	Labels      map[string]string `yaml:"labels"`
	Automations []string          `yaml:"automations"`
	Buttons     []string          `yaml:"buttons"`
}
