package zagreus

import (
	"context"
	"log/slog"
	"time"

	"github.com/woodleighschool/zagreus/internal/app"
)

func runLoop(ctx context.Context, interval time.Duration, service *app.Service, wake <-chan struct{}, logger *slog.Logger) {
	runCycle(ctx, service.Sync, logger)
	for ctx.Err() == nil {
		delay := interval
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-wake:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		case <-timer.C:
		}
		runCycle(ctx, service.Sync, logger)
	}
}

func runCycle(ctx context.Context, sync func(context.Context) ([]app.Result, error), logger *slog.Logger) {
	started := time.Now()
	results, err := sync(ctx)
	if ctx.Err() != nil {
		return
	}
	for _, result := range results {
		attributes := []any{
			"title", result.Title,
			"action", result.Action,
			"reason", result.Reason,
		}
		if result.Error != nil {
			logger.ErrorContext(ctx, "export failed", append(attributes, "error", result.Error))
		} else {
			logger.DebugContext(ctx, "export succeeded", attributes)
		}
	}
	if err != nil {
		logger.ErrorContext(ctx, "sync cycle failed", "vulns", len(results), "duration", time.Since(started), "error", err)
		return
	}
	logger.InfoContext(ctx, "sync cycle succeeded", "vulns", len(results), "duration", time.Since(started))
}
