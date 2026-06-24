package optimizer

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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

// dockerInspectResult mirrors the fields we care about from `docker inspect`.
type dockerInspectResult struct {
	ID           string `json:"Id"`
	Size         int64  `json:"Size"`
	Architecture string `json:"Architecture"`
	Os           string `json:"Os"`
	RootFS       struct {
		Layers []string `json:"Layers"`
	} `json:"RootFS"`
}

// InspectImage calls `docker inspect` and returns size metadata for the given
// image reference. Requires the Docker CLI to be available in PATH.
func InspectImage(ctx context.Context, imageRef string) (*ImageStats, error) {
	out, err := exec.CommandContext(ctx, "docker", "inspect",
		"--type", "image",
		imageRef,
	).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("docker inspect %q: %s", imageRef, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("docker inspect %q: %w", imageRef, err)
	}

	var results []dockerInspectResult
	if err := json.Unmarshal(out, &results); err != nil {
		return nil, fmt.Errorf("parse docker inspect output: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("image %q not found", imageRef)
	}

	r := results[0]
	return &ImageStats{
		Ref:          imageRef,
		ID:           shortID(r.ID),
		SizeBytes:    r.Size,
		Layers:       len(r.RootFS.Layers),
		Architecture: r.Architecture,
		OS:           r.Os,
	}, nil
}

// ComparisonResult holds the before/after measurement and computed reduction.
type ComparisonResult struct {
	Before       *ImageStats
	After        *ImageStats
	SavedBytes   int64
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

// FormatBytes is the exported variant for use in cmd packages.
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
