package vault

import (
	"context"
	"time"
)

// WatchOptions configures secret watching behavior.
type WatchOptions struct {
	Interval time.Duration
	MaxDrift int // max version drift before alerting
}

// VersionChange represents a detected change between polls.
type VersionChange struct {
	Path       string
	FromVersion int
	ToVersion   int
	DetectedAt  time.Time
}

// WatchSecret polls a secret path at the given interval and emits VersionChange
// events when the latest version advances.
func WatchSecret(ctx context.Context, c *Client, path string, opts WatchOptions, out chan<- VersionChange) error {
	versions, err := ListVersions(c, path)
	if err != nil {
		return err
	}
	last := latestVersion(versions)

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			versions, err = ListVersions(c, path)
			if err != nil {
				continue
			}
			current := latestVersion(versions)
			if current > last {
				out <- VersionChange{
					Path:        path,
					FromVersion: last,
					ToVersion:   current,
					DetectedAt:  time.Now().UTC(),
				}
				last = current
			}
		}
	}
}

func latestVersion(versions []int) int {
	max := 0
	for _, v := range versions {
		if v > max {
			max = v
		}
	}
	return max
}
