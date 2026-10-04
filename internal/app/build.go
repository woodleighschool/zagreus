package app

import (
	"fmt"
	"time"

	"github.com/woodleighschool/zagreus/external/nessus"
	"github.com/woodleighschool/zagreus/external/trello"
	"github.com/woodleighschool/zagreus/internal/config"
)

type Service struct {
	config *config.Config
	nessus *nessus.Client
	trello *trello.Client

	now func() time.Time
}

func New(cfg *config.Config) (*Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	nessusClient, err := nessus.NewClient(nessus.Config{
		Host:      cfg.Nessus.Connection.Host,
		AccessKey: cfg.Nessus.Connection.AccessKey,
		SecretKey: cfg.Nessus.Connection.SecretKey,
	})
	if err != nil {
		return nil, fmt.Errorf("app startup: nessus: %w", err)
	}
	trelloClient, err := trello.NewClient(trello.Config{
		AccessKey:   cfg.Trello.Connection.APIKey,
		AccessToken: cfg.Trello.Connection.APIToken,
	})
	if err != nil {
		return nil, fmt.Errorf("app startup: trello: %w", err)
	}
	return &Service{
		config: cfg,
		nessus: nessusClient,
		trello: trelloClient,
		now:    time.Now,
	}, nil
}
