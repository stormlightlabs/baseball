package seed

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultRetrosheetRequestInterval = 400 * time.Millisecond

var (
	retrosheetThrottleMu  sync.Mutex
	retrosheetNextAllowed time.Time
)

// retrosheetRequestInterval returns the minimum spacing between outbound
// Retrosheet HTTP requests. Set RETROSHEET_MIN_REQUEST_INTERVAL_MS to override.
func retrosheetRequestInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv("RETROSHEET_MIN_REQUEST_INTERVAL_MS"))
	if raw == "" {
		return defaultRetrosheetRequestInterval
	}

	ms, err := strconv.Atoi(raw)
	if err != nil || ms < 0 {
		return defaultRetrosheetRequestInterval
	}
	return time.Duration(ms) * time.Millisecond
}

// waitForRetrosheetDownloadSlot enforces process-wide request spacing so even
// parallel workers don't burst traffic toward Retrosheet.
func waitForRetrosheetDownloadSlot(ctx context.Context) error {
	interval := retrosheetRequestInterval()
	if interval <= 0 {
		return nil
	}

	retrosheetThrottleMu.Lock()
	now := time.Now()
	slot := now
	if retrosheetNextAllowed.After(now) {
		slot = retrosheetNextAllowed
	}
	retrosheetNextAllowed = slot.Add(interval)
	retrosheetThrottleMu.Unlock()

	wait := time.Until(slot)
	if wait <= 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

