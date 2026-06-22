package optimizer

import "testing"

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
