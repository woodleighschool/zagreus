package zagreus

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/woodleighschool/zagreus/internal/app"
	"github.com/woodleighschool/zagreus/internal/config"
)

func newRootCommand() *cobra.Command {
	var configPaths []string
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
	command.PersistentFlags().StringArrayVar(
		&configPaths,
		"config",
		defaultConfigPaths(),
		"path to a YAML configuration file; may be repeated in overlay order",
	)
	command.AddCommand(
		newRunCommand(&configPaths),
		newSyncCommand(&configPaths),
		newSchemaCommand(),
		newVersionCommand(),
	)
	return command
}

func defaultConfigPaths() []string {
	info, err := os.Stat("config.yaml")
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	return []string{"config.yaml"}
}

func newRunCommand(configPaths *[]string) *cobra.Command {
	var once bool
	var logLevel string
	command := &cobra.Command{
		Use:   "run",
		Short: "Syncs vulnerabilities from Nessus to Trello immediately, and then again at configured intervals",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			level, err := parseLogLevel(logLevel)
			if err != nil {
				return err
			}
			application, interval, err := buildOperationalApp(*configPaths)
			if err != nil {
				return err
			}
			logger := slog.New(slog.NewJSONHandler(command.ErrOrStderr(), &slog.HandlerOptions{Level: level}))
			logger.Info("Zagreus started", "version", version, "config", *configPaths, "once", once)
			syncDone := make(chan struct{})
			wake := make(chan struct{}, 1)
			go func() {
				runLoop(command.Context(), interval, application, wake, logger)
				close(syncDone)
			}()
			<-syncDone
			return nil
		},
	}
	command.Flags().BoolVar(&once, "once", false, "run once and exit")
	command.Flags().StringVar(&logLevel, "log-level", "info", "log level: debug, info, warn, or error")
	return command
}

func newSyncCommand(configPaths *[]string) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Syncs all vulnerabilities from latest Nessus report to Trello on demand",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			application, _, err := buildOperationalApp(*configPaths)
			if err != nil {
				return err
			}
			results, err := application.Sync(command.Context())
			return errors.Join(writeSyncResults(command.OutOrStdout(), results), err)
		},
	}
}

func newSchemaCommand() *cobra.Command {
	var outputPath string
	command := &cobra.Command{
		Use:   "schema",
		Short: "Generate configuration file schema",
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
			if err := os.WriteFile(outputPath, document, 0o644); err != nil {
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
			_, err := fmt.Fprintln(command.OutOrStdout(), "zagreus %s\ncommit: %s\nbuilt %s\n", version, commit, date)
			return err
		},
	}
}

func buildOperationalApp(configPaths []string) (*app.Service, time.Duration, error) {
	cfg, err := config.Load(configPaths...)
	if err != nil {
		return nil, time.Microsecond, fmt.Errorf("load configuration: %w", err)
	}
	application, err := app.New(cfg)
	if err != nil {
		return nil, time.Microsecond, fmt.Errorf("start service: %w", err)
	}
	return application, cfg.Settings.Sync.PollInterval.Duration, nil
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
