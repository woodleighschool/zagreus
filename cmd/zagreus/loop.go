package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/woodleighschool/zagreus/internal/app"
)

type service interface {
	Sync(context.Context) ([]app.Result, error)
}

func runLoop(ctx context.Context, interval time.Duration, service service, logger *slog.Logger) {
	for {
		_ = runCycle(ctx, service, logger)
		if ctx.Err() != nil {
			return
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}
	}
}

func runCycle(ctx context.Context, service service, logger *slog.Logger) error {
	started := time.Now()
	results, err := service.Sync(ctx)
	if ctx.Err() != nil {
		return err
	}
	for _, result := range results {
		logResult(logger, result)
	}
	attributes := []any{
		"cards", len(results),
		"created", countOutcome(results, app.OutcomeCreated),
		"updated", countOutcome(results, app.OutcomeUpdated),
		"closed", countOutcome(results, app.OutcomeClosed),
		"reopened", countOutcome(results, app.OutcomeReopened),
		"duration", time.Since(started),
	}
	if err != nil {
		logger.ErrorContext(ctx, "sync failed", append(attributes, "failed", countOutcome(results, app.OutcomeFailed), "error", err)...)
	} else {
		logger.InfoContext(ctx, "sync complete", attributes...)
	}
	return err
}

func logResult(logger *slog.Logger, result app.Result) {
	attributes := []any{
		"title", result.Title,
		"action", result.Action,
		"outcome", result.Outcome,
		"reason", result.Reason,
	}
	if result.Vulnerability != nil {
		attributes = append(attributes, "labels", result.Vulnerability.Labels)
		attributes = append(attributes, "hosts", len(result.Vulnerability.Hosts))
	}
	if result.Error != nil {
		attributes = append(attributes, "error", *result.Error)
	}
	switch result.Outcome {
	case app.OutcomeCreated:
		logger.Info("created card", attributes...)
	case app.OutcomeUpdated:
		logger.Info("updated card", attributes...)
	case app.OutcomeReopened:
		logger.Info("reopened card", attributes...)
	case app.OutcomeClosed:
		logger.Info("closed card", attributes...)
	case app.OutcomeSkipped:
		logger.Debug("skipped card", attributes...)
	case app.OutcomeFailed:
		logger.Warn("failed to change card", attributes...)
	case app.OutcomePlan:
		attributes = append(attributes, "action", result.Action)
		logger.Warn("card planned", attributes...)
	}
}

func countOutcome(results []app.Result, outcome app.Outcome) int {
	count := 0
	for _, result := range results {
		if result.Outcome == outcome {
			count++
		}
	}
	return count
}
