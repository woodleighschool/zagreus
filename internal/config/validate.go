package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TODO: Update

const supportedVersion = 1

func (c *Config) applyDefaults() {
	if c.Settings.LogLevel == "" {
		c.Settings.LogLevel = "info"
	}
	if !c.Settings.Sync.PollInterval.set {
		c.Settings.Sync.PollInterval.Duration = time.Hour
	}
	if !c.Settings.Sync.RetryAfter.set {
		c.Settings.Sync.RetryAfter.Duration = 5 * time.Minute
	}
	if !c.Settings.Sync.RetryMax.set {
		c.Settings.Sync.RetryMax.Duration = 30 * time.Minute
	}
}

func (c *Config) validate() error {
	if c.Version != supportedVersion {
		return fmt.Errorf("config version must be %d, found %d", supportedVersion, c.Version)
	}
	c.applyDefaults()
	if err := c.validateSettings(); err != nil {
		return err
	}
	if err := c.validateNessus(); err != nil {
		return err
	}
	return c.validateTrello()
}

func (c *Config) validateSettings() error {
	if err := c.validateLogLevel(); err != nil {
		return err
	}
	if err := c.validateSyncSettings(); err != nil {
		return err
	}
	return c.validateMode()
}

func (c *Config) validateLogLevel() error {
	switch strings.ToLower(c.Settings.LogLevel) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("log level must be debug|info|warn|error, got %s", c.Settings.LogLevel)
	}
}

func (c *Config) validateSyncSettings() error {
	sync := c.Settings.Sync
	if sync.PollInterval.Duration < 5*time.Minute {
		return fmt.Errorf("application_settings.sync.poll_interval must not be less than 5 minutes")
	}
	if sync.RetryAfter.Duration > sync.PollInterval.Duration ||
		sync.RetryAfter.Duration < time.Minute {
		return fmt.Errorf("application_settings.sync.retry_after must be longer than a minute and less than application_settings.sync.poll_interval")
	}
	if sync.RetryMax.Duration > 2*time.Hour || sync.RetryMax.Duration < 3*sync.RetryAfter.Duration {
		return fmt.Errorf("application_settings.sync.retry_max must be less than 2 hours and greater than 3 x application_settings.sync.retry_after")
	}
	return nil
}

func (c *Config) validateMode() error {
	switch strings.ToLower(c.Settings.Mode) {
	case "create", "move":
		return nil
	default:
		return fmt.Errorf("settings.mode must be create|move got %s", c.Settings.Mode)
	}
}

func (c *Config) validateNessus() error {
	if err := c.validateNessusConnection(); err != nil {
		return err
	}
	return c.validateNessusSettings()
}

func (c *Config) validateNessusConnection() error {
	host := c.Nessus.Connection.Host
	parsed, err := url.Parse(host)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("nessus.connection.host is required: https://{your_host}")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("nessus.connection.host must not contain query or fragment")
	}

	if c.Nessus.Connection.AccessKey == "" {
		return fmt.Errorf("nessus.connection.access_key is required")
	}
	if c.Nessus.Connection.SecretKey == "" {
		return fmt.Errorf("nessus.connection.secret_key is required")
	}
	return nil
}

func (c *Config) validateNessusSettings() error {
	for index, ignoredSeverity := range c.Nessus.Settings.Ignore.Severity {
		if err := validateSeverity(ignoredSeverity); err != nil {
			return fmt.Errorf("nessus.settings.ignore.severity[%d]: %w", index, err)
		}
	}
	for hostname, settings := range c.Nessus.Settings.Ignore.Hosts {
		if err := validateHost(settings); err != nil {
			return fmt.Errorf("nessus.settings.ignore.hosts[%s]: %w", hostname, err)
		}
	}
	return nil
}

func validateSeverity(severity string) error {
	switch severity {
	case "info", "low", "medium", "high", "critical":
		return nil
	default:
		return fmt.Errorf("must be info|low|medium|high|critical, got %s", severity)
	}
}
func validateHost(host NessusHostExclusion) error {
	if host.Full == nil && len(host.Plugins) == 0 && len(host.Severity) == 0 {
		return fmt.Errorf("must contain at least one condition")
	}
	if host.Full != nil && *host.Full {
		if len(host.Plugins) != 0 || len(host.Severity) != 0 {
			return fmt.Errorf("cannot be fully excluded and have severity/plugin rules")
		}
	}
	return nil
}

func (c *Config) validateTrello() error {
	if err := c.validateTrelloConnection(); err != nil {
		return err
	}
	return c.validateTrelloSettings()
}

func (c *Config) validateTrelloConnection() error {
	if c.Trello.Connection.APIKey == "" {
		return fmt.Errorf("trello.connection.api_key is required")
	}
	if c.Trello.Connection.APIToken == "" {
		return fmt.Errorf("trello.connection.api_token is required")
	}
	return nil
}

func (c *Config) validateTrelloSettings() error {
	// TODO: Better validation?
	return nil
}
