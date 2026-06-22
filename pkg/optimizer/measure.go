package optimizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
)

// ImageStats holds size information retrieved from the Docker daemon.
type ImageStats struct {
	Ref          string
	ID           string
	SizeBytes    int64
	Layers       int
	Architecture string
	OS           string
}

// HumanSize returns a human-readable size string (KB / MB / GB).
func (s *ImageStats) HumanSize() string {
	return humanBytes(s.SizeBytes)
}

// InspectImage connects to the local Docker daemon and returns size metadata
// for the given image reference (e.g. "myapp:latest" or a digest).
func InspectImage(ctx context.Context, imageRef string) (*ImageStats, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to Docker daemon: %w", err)
	}
	defer cli.Close()

	info, _, err := cli.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("inspect image %q: %w", imageRef, err)
	}

	return &ImageStats{
		Ref:          imageRef,
		ID:           shortID(info.ID),
		SizeBytes:    info.Size,
		Layers:       len(info.RootFS.Layers),
		Architecture: info.Architecture,
		OS:           info.Os,
	}, nil
}

// CompareImages measures two images and returns the reduction statistics.
type ComparisonResult struct {
	Before      *ImageStats
	After       *ImageStats
	SavedBytes  int64
	ReductionPct float64
}

// CompareImages inspects both images and computes the size delta.
func CompareImages(ctx context.Context, beforeRef, afterRef string) (*ComparisonResult, error) {
	before, err := InspectImage(ctx, beforeRef)
	if err != nil {
		return nil, fmt.Errorf("before image: %w", err)
	}
	after, err := InspectImage(ctx, afterRef)
	if err != nil {
		return nil, fmt.Errorf("after image: %w", err)
	}

	saved := before.SizeBytes - after.SizeBytes
	var pct float64
	if before.SizeBytes > 0 {
		pct = float64(saved) / float64(before.SizeBytes) * 100
	}

	return &ComparisonResult{
		Before:       before,
		After:        after,
		SavedBytes:   saved,
		ReductionPct: pct,
	}, nil
}

// FormatBytes is the exported variant of humanBytes for use in cmd packages.
func FormatBytes(b int64) string { return humanBytes(b) }

func humanBytes(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func shortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
