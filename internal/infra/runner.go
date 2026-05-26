package infra

import (
	"context"
	"log/slog"
	"time"

	"github.com/ckinan/lab/internal/domain"
)

func Start(ctx context.Context, c *domain.Collector, interval time.Duration) <-chan domain.Snapshot {
	ch := make(chan domain.Snapshot, 1)

	go func() {
		defer close(ch)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		if snap, err := c.Collect(); err == nil {
			ch <- snap
		}

		for {
			select {
			case <-ctx.Done():
				return // context canceled
			case <-ticker.C:
				snap, err := c.Collect()
				if err != nil {
					slog.Warn("collect error", "err", err)
					continue // skip this tick, try again next interval
				}
				select {
				case ch <- snap:
					// sent successfully
				default:
					// UI hasn't consumed the previous snapshot yet - drop this one
					// better to drop a tick than to block the goroutine
				}
			}
		}
	}()

	return ch
}
