package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/woodleighschool/zagreus/internal/app"
	"github.com/woodleighschool/zagreus/internal/config"
)

type cli struct {
	configPaths []string
	output      *commandOutput
}

func newRootCommand() (*cobra.Command, *commandOutput) {
	c := &cli{output: &commandOutput{}}
	command := &cobra.Command{
		Use:           "zagreus",
		Short:         "Syncs current vulnerabilities from Nessus Professional into Trello",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
	output := c.output
	command.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		return output.start(cmd)
	}
	command.SetVersionTemplate(fmt.Sprintf("zagreus %s (commit %s, built %s)\n", version, commit, date))
	command.PersistentFlags().Bool("no-progress", false, "Disable terminal progress")
	command.PersistentFlags().StringArrayVar(
		&c.configPaths,
		"config",
		defaultConfigPaths(),
		"path to a YAML configuration file; may be repeated in overlay mode",
	)
	command.PersistentFlags().StringVar(&c.output.level, "log-level", "info", "Override the config.yaml log-level value")
	command.AddCommand(
		c.validateCommand(),
		c.planCommand(),
		c.syncCommand(),
		c.runCommand(),
		newSchemaCommand(),
		newVersionCommand(),
	)
	return command, output
}

func defaultConfigPaths() []string {
	info, err := os.Stat("config.yaml")
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	return []string{"config.yaml"}
}

func (c *cli) validateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration and Trello board state",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, err := c.loadConfig(command)
			if err != nil {
				return fmt.Errorf("validate configuration: %w", err)
			}
			service, err := app.New(cfg, c.output.logger)
			if err != nil {
				return fmt.Errorf("start service: %w", err)
			}
			if err := service.Validate(command.Context()); err != nil {
				_, err := fmt.Fprintln(command.OutOrStdout(), newTextStyle(command.OutOrStdout()).paint("✗ Failed to validate environment", color.FgHiRed))
				return err
			}
			_, err = fmt.Fprintln(command.OutOrStdout(), newTextStyle(command.OutOrStdout()).paint("✓ Configuration/State is valid", color.FgHiGreen))
			return err
		},
	}
}

func (c *cli) planCommand() *cobra.Command {
	var includeUnchanged bool
	var jsonOutput bool
	command := &cobra.Command{
		Use:   "plan",
		Short: "Print out a plan of what actions would be commited if a sync was run",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, err := c.loadConfig(command)
			if err != nil {
				return err
			}
			service, err := app.New(cfg, c.output.logger)
			if err != nil {
				return fmt.Errorf("start service: %w", err)
			}
			plans, planErr := service.Plan(command.Context(), includeUnchanged)
			c.output.stop()
			if planErr != nil {
				return planErr
			}
			if plans == nil {
				c.output.logger.Info("Nothing to action")
			}
			return writeReport(command.OutOrStdout(), jsonOutput, includeUnchanged, false, plans, planErr)
		},
	}
	command.Flags().BoolVar(&includeUnchanged, "include-unchanged", false, "Include unchanged cards in the plan report")
	command.Flags().BoolVar(&jsonOutput, "json", false, "Write the plan report as JSON")
	return command
}

func (c *cli) syncCommand() *cobra.Command {
	var jsonOutput bool
	var includeUnchanged bool
	command := &cobra.Command{
		Use:   "sync",
		Short: "Sync from Nessus to Trello once, and then print the results",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, err := c.loadConfig(command)
			if err != nil {
				return err
			}
			service, err := app.New(cfg, c.output.logger)
			if err != nil {
				return fmt.Errorf("start service: %w", err)
			}
			results, resultErr := service.Sync(command.Context())
			c.output.stop()
			if resultErr != nil {
				return resultErr
			}
			if results == nil {
				c.output.logger.Info("Nothing to action")
				return nil
			}
			return writeReport(command.OutOrStdout(), jsonOutput, includeUnchanged, true, results, resultErr)
		},
	}
	command.Flags().BoolVar(&jsonOutput, "json", false, "Write the report as JSON")
	command.Flags().BoolVar(&includeUnchanged, "include-unchanged", false, "Include unchanged cards in report")
	return command
}

func (c *cli) runCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Sync immediately, then again at scheduled intervals",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, err := c.loadConfig(command)
			if err != nil {
				return err
			}
			service, err := app.New(cfg, c.output.logger)
			if err != nil {
				return fmt.Errorf("start service: %w", err)
			}
			c.output.logger.Info("Service started", "version", version)
			runLoop(command.Context(), cfg.Settings.Sync.PollInterval.Duration, service, c.output.logger)
			return nil
		},
	}
}

func newSchemaCommand() *cobra.Command {
	var outputPath string
	command := &cobra.Command{
		Use:   "schema",
		Short: "Generate the JSON schema used by YAML editors",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			document, err := config.JSONSchemaDocument()
			if err != nil {
				return fmt.Errorf("generate config schema: %w", err)
			}
			if outputPath == "-" {
				_, err = command.OutOrStdout().Write(document)
				return err
			}
			if err := os.WriteFile(outputPath, document, 0o644); err != nil { // #nosec G306 - Containerized, we always control the file system
				return fmt.Errorf("write config schema: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringVar(&outputPath, "output", "-", "schema output path, or - for stdout")
	return command
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(command.OutOrStdout(), "zagreus %s (commit %s, built %s)\n", version, commit, date)
			return err
		},
	}
}

func (c *cli) loadConfig(command *cobra.Command) (*config.Config, error) {
	cfg, err := config.Load(c.configPaths...)
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}
	if c.output.daemon && !command.Flags().Changed("log-level") {
		level, err := parseLogLevel(cfg.Settings.LogLevel)
		if err != nil {
			return nil, err
		}
		c.output.threshold.Set(level)
	}
	c.output.logger.DebugContext(command.Context(), "Loading application")
	return cfg, nil
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("log level must be debug|info|warn|error, received %s", value)
	}
}
