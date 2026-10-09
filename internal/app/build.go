package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/woodleighschool/zagreus/external/nessus"
	"github.com/woodleighschool/zagreus/external/trello"
	"github.com/woodleighschool/zagreus/internal/config"
)

type Service struct {
	config *config.Config
	logger *slog.Logger

	runtime RuntimeSettings

	nessus *nessus.Client
	trello *trello.Client

	now func() time.Time
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	logger.DebugContext(ctx, "creating nessus client")
	nessusClient, err := nessus.NewClient(nessus.Config{
		Host:      cfg.Nessus.Connection.Host,
		AccessKey: cfg.Nessus.Connection.AccessKey,
		SecretKey: cfg.Nessus.Connection.SecretKey,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create nessus client", "err", err)
		return nil, fmt.Errorf("app startup: nessus: %w", err)
	}
	logger.DebugContext(ctx, "nessus client created")
	logger.DebugContext(ctx, "creating trello client")
	trelloClient, err := trello.NewClient(trello.Config{
		AccessKey:   cfg.Trello.Connection.APIKey,
		AccessToken: cfg.Trello.Connection.APIToken,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create trello client", "err", err)
		return nil, fmt.Errorf("app startup: trello: %w", err)
	}
	logger.DebugContext(ctx, "trello client created")
	logger.DebugContext(ctx, "finished app init")
	return &Service{
		config: cfg,
		logger: logger,
		nessus: nessusClient,
		trello: trelloClient,
		now:    time.Now,
	}, nil
}
