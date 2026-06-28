package optimizer

import (
	"os"
	"testing"
)

const singleStageDockerfile = `FROM ubuntu:22.04
RUN apt-get update && apt-get install -y curl
COPY . /app
CMD ["/app/server"]
`

const multiStageDockerfile = `FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /bin/server .

FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /bin/server /server
USER nobody
ENTRYPOINT ["/server"]
`

func TestParse_SingleStage(t *testing.T) {
	df := ParseString(singleStageDockerfile)

	if df.IsMultiStage {
		t.Error("expected single-stage, got multi-stage")
	}
	if df.BaseImage != "ubuntu:22.04" {
		t.Errorf("expected base image ubuntu:22.04, got %s", df.BaseImage)
	}
	if df.PackageManager != "apt" {
		t.Errorf("expected apt package manager, got %s", df.PackageManager)
	}
}

func TestParse_MultiStage(t *testing.T) {
	df := ParseString(multiStageDockerfile)

	if !df.IsMultiStage {
		t.Error("expected multi-stage")
	}
	if len(df.Stages) != 2 {
		t.Errorf("expected 2 stages, got %d", len(df.Stages))
	}
	if df.Stages[0].Name != "builder" {
		t.Errorf("expected first stage name 'builder', got %q", df.Stages[0].Name)
	}
	if !df.HasDistroless {
		t.Error("expected distroless base detected")
	}
	if !df.HasNobodyUser {
		t.Error("expected nobody user detected")
	}
}

func TestAnalyzeIssues_SingleStage(t *testing.T) {
	df := ParseString(singleStageDockerfile)
	issues := AnalyzeIssues(df)

	if len(issues) == 0 {
		t.Error("expected issues for naive single-stage Dockerfile")
	}

	hasMultiStageIssue := false
	for _, i := range issues {
		if i == "Single-stage build detected: consider multi-stage to separate build and runtime" {
			hasMultiStageIssue = true
		}
	}
	if !hasMultiStageIssue {
		t.Error("expected multi-stage issue in analysis")
	}
}

func TestAnalyzeIssues_OptimizedDockerfile(t *testing.T) {
	df := ParseString(multiStageDockerfile)
	issues := AnalyzeIssues(df)

	if len(issues) != 0 {
		t.Errorf("expected no issues for optimized Dockerfile, got: %v", issues)
	}
}

func TestParseFile(t *testing.T) {
	content := `FROM ubuntu:22.04
RUN echo hello
`
	tmpfile, err := os.CreateTemp("", "Dockerfile.*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	df, err := ParseFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if df.BaseImage != "ubuntu:22.04" {
		t.Errorf("expected BaseImage ubuntu:22.04, got %s", df.BaseImage)
	}

	_, err = ParseFile("non-existent-file-path-xyz")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestParsePlatformOption(t *testing.T) {
	content := `FROM --platform=linux/amd64 ubuntu:20.04 AS builder
RUN echo building

FROM --platform=linux/arm64 gcr.io/distroless/static-debian12:latest
COPY --from=builder /bin/server /server
`
	df := ParseString(content)
	if df.BaseImage != "ubuntu:20.04" {
		t.Errorf("expected BaseImage ubuntu:20.04, got %s", df.BaseImage)
	}
	if len(df.Stages) != 2 {
		t.Errorf("expected 2 stages, got %d", len(df.Stages))
	}
	if df.Stages[0].Image != "ubuntu:20.04" {
		t.Errorf("expected stage 0 image ubuntu:20.04, got %s", df.Stages[0].Image)
	}
	if df.Stages[0].Name != "builder" {
		t.Errorf("expected stage 0 name builder, got %s", df.Stages[0].Name)
	}
	if df.Stages[1].Image != "gcr.io/distroless/static-debian12:latest" {
		t.Errorf("expected stage 1 image distroless, got %s", df.Stages[1].Image)
	}
}

func TestParseUserDeclarationsSpacing(t *testing.T) {
	cases := []string{
		"FROM alpine\nUSER\tnobody\n",
		"FROM alpine\nUSER    nobody\n",
		"FROM alpine\nUSER \t nobody\n",
		"FROM alpine\nUSER 65534\n",
		"FROM alpine\nUSER \t 65534\n",
	}

	for i, c := range cases {
		df := ParseString(c)
		if !df.HasNobodyUser {
			t.Errorf("case %d: expected nobody user to be detected in %q", i, c)
		}
	}
}

func TestParseInvalidStatementsAndEmpty(t *testing.T) {
	df := ParseString("")
	if df.BaseImage != "" {
		t.Errorf("expected empty base image, got %q", df.BaseImage)
	}
	if len(df.Stages) != 0 {
		t.Errorf("expected 0 stages, got %d", len(df.Stages))
	}

	df2 := ParseString("FROM ubuntu AS")
	if len(df2.Stages) != 1 {
		t.Errorf("expected 1 stage, got %d", len(df2.Stages))
	}
	if df2.Stages[0].Image != "ubuntu" {
		t.Errorf("expected image ubuntu, got %q", df2.Stages[0].Image)
	}
	if df2.Stages[0].Name != "" {
		t.Errorf("expected empty stage name, got %q", df2.Stages[0].Name)
	}

	df3 := ParseString("FROM --platform=linux/amd64")
	if len(df3.Stages) != 0 {
		t.Errorf("expected 0 stages for dangling platform FROM, got %d", len(df3.Stages))
	}
}
