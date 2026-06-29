package optimizer

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type mockRunner struct {
	runFunc func(ctx context.Context, name string, arg ...string) ([]byte, error)
}

func (m *mockRunner) Run(ctx context.Context, name string, arg ...string) ([]byte, error) {
	return m.runFunc(ctx, name, arg...)
}

type mockExitErr struct {
	msg    string
	stderr []byte
}

func (e *mockExitErr) Error() string   { return e.msg }
func (e *mockExitErr) Stderr() []byte { return e.stderr }

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		input int64
		want  string
	}{
		{500, "500 B"},
		{2048, "2.00 KB"},
		{5 * 1024 * 1024, "5.00 MB"},
		{2 * 1024 * 1024 * 1024, "2.00 GB"},
	}
	for _, c := range cases {
		got := humanBytes(c.input)
		if got != c.want {
			t.Errorf("humanBytes(%d) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestShortID(t *testing.T) {
	full := "sha256:abc123def456789xyz"
	got := shortID(full)
	if got != "abc123def456" {
		t.Errorf("shortID(%q) = %q, want %q", full, got, "abc123def456")
	}

	plain := "deadbeef0011"
	if shortID(plain) != plain {
		t.Errorf("shortID of short string should be unchanged")
	}
}

func TestComparisonResult_ReductionPct(t *testing.T) {
	r := &ComparisonResult{
		Before:     &ImageStats{SizeBytes: 100 * 1024 * 1024},
		After:      &ImageStats{SizeBytes: 10 * 1024 * 1024},
		SavedBytes: 90 * 1024 * 1024,
	}
	if r.Before.SizeBytes > 0 {
		r.ReductionPct = float64(r.SavedBytes) / float64(r.Before.SizeBytes) * 100
	}
	if r.ReductionPct != 90.0 {
		t.Errorf("expected 90%% reduction, got %.2f%%", r.ReductionPct)
	}
}

func TestInspectImage_Success(t *testing.T) {
	oldRunner := Runner
	defer func() { Runner = oldRunner }()

	mockJSON := `[
		{
			"Id": "sha256:abc123def456789xyz",
			"Size": 102400,
			"Architecture": "amd64",
			"Os": "linux",
			"RootFS": {
				"Layers": [
					"sha256:layer1",
					"sha256:layer2"
				]
			}
		}
	]`

	Runner = &mockRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			if name != "docker" || arg[0] != "inspect" || arg[1] != "--type" || arg[2] != "image" || arg[3] != "my-image:latest" {
				t.Errorf("unexpected command: %s %v", name, arg)
			}
			return []byte(mockJSON), nil
		},
	}

	stats, err := InspectImage(context.Background(), "my-image:latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Ref != "my-image:latest" {
		t.Errorf("expected Ref %q, got %q", "my-image:latest", stats.Ref)
	}
	if stats.ID != "abc123def456" {
		t.Errorf("expected ID %q, got %q", "abc123def456", stats.ID)
	}
	if stats.SizeBytes != 102400 {
		t.Errorf("expected SizeBytes %d, got %d", 102400, stats.SizeBytes)
	}
	if stats.Layers != 2 {
		t.Errorf("expected Layers %d, got %d", 2, stats.Layers)
	}
	if stats.Architecture != "amd64" {
		t.Errorf("expected Architecture %q, got %q", "amd64", stats.Architecture)
	}
	if stats.OS != "linux" {
		t.Errorf("expected OS %q, got %q", "linux", stats.OS)
	}
}

func TestInspectImage_MissingImage(t *testing.T) {
	oldRunner := Runner
	defer func() { Runner = oldRunner }()

	Runner = &mockRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return nil, &mockExitErr{
				msg:    "exit status 1",
				stderr: []byte("Error: No such image: missing-image:latest\n"),
			}
		},
	}

	stats, err := InspectImage(context.Background(), "missing-image:latest")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := `docker inspect "missing-image:latest": Error: No such image: missing-image:latest`
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
	if stats != nil {
		t.Errorf("expected nil stats, got %+v", stats)
	}
}

func TestInspectImage_MalformedJSON(t *testing.T) {
	oldRunner := Runner
	defer func() { Runner = oldRunner }()

	Runner = &mockRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(`{invalid json`), nil
		},
	}

	_, err := InspectImage(context.Background(), "malformed-image")
	if err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse docker inspect output") {
		t.Errorf("expected JSON parse error, got %v", err)
	}
}

func TestInspectImage_EmptyJSON(t *testing.T) {
	oldRunner := Runner
	defer func() { Runner = oldRunner }()

	Runner = &mockRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(`[]`), nil
		},
	}

	_, err := InspectImage(context.Background(), "empty-image")
	if err == nil {
		t.Fatal("expected error on empty JSON array, got nil")
	}
	expectedErr := `image "empty-image" not found`
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestCompareImages(t *testing.T) {
	oldRunner := Runner
	defer func() { Runner = oldRunner }()

	mockJSONBefore := `[{"Id": "sha256:before123", "Size": 200, "Architecture": "amd64", "Os": "linux", "RootFS": {"Layers": []}}]`
	mockJSONAfter := `[{"Id": "sha256:after123", "Size": 50, "Architecture": "amd64", "Os": "linux", "RootFS": {"Layers": []}}]`

	Runner = &mockRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			img := arg[3]
			if img == "before" {
				return []byte(mockJSONBefore), nil
			}
			if img == "after" {
				return []byte(mockJSONAfter), nil
			}
			return nil, fmt.Errorf("unknown image")
		},
	}

	res, err := CompareImages(context.Background(), "before", "after")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.SavedBytes != 150 {
		t.Errorf("expected 150 saved bytes, got %d", res.SavedBytes)
	}
	if res.ReductionPct != 75.0 {
		t.Errorf("expected 75%% reduction, got %.2f%%", res.ReductionPct)
	}
}

func TestImageStats_HumanSizeAndFormatBytes(t *testing.T) {
	stats := &ImageStats{SizeBytes: 1536}
	if stats.HumanSize() != "1.50 KB" {
		t.Errorf("expected 1.50 KB, got %s", stats.HumanSize())
	}
	if FormatBytes(2048) != "2.00 KB" {
		t.Errorf("expected 2.00 KB, got %s", FormatBytes(2048))
	}
}
