package optimizer

import (
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
