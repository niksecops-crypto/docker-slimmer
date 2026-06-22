package optimizer

import (
	"strings"
	"testing"
)

func TestOptimize_ContainsBuilderStage(t *testing.T) {
	opt := NewOptimizer("golang:1.22-alpine")
	out := opt.Optimize([]string{"/app/server"}, nil)

	if !strings.Contains(out, "AS builder") {
		t.Error("expected multi-stage builder")
	}
	if !strings.Contains(out, "distroless") {
		t.Error("expected distroless runtime image")
	}
}

func TestOptimize_ArtifactsCopied(t *testing.T) {
	opt := NewOptimizer("golang:1.22-alpine")
	artifacts := []string{"/app/server", "/app/config"}
	out := opt.Optimize(artifacts, nil)

	for _, a := range artifacts {
		if !strings.Contains(out, a) {
			t.Errorf("expected artifact %q in output", a)
		}
	}
}

func TestOptimize_EnvVars(t *testing.T) {
	opt := NewOptimizer("golang:1.22-alpine")
	out := opt.Optimize([]string{"/app/server"}, map[string]string{"GIN_MODE": "release"})

	if !strings.Contains(out, "GIN_MODE") {
		t.Error("expected env var in output")
	}
}

func TestCleanCommands_Apt(t *testing.T) {
	opt := NewOptimizer("ubuntu:22.04")
	cmd := opt.CleanCommands("apt")
	if !strings.Contains(cmd, "apt-get") {
		t.Error("expected apt-get in clean command")
	}
	if !strings.Contains(cmd, "rm -rf /var/lib/apt/lists/*") {
		t.Error("expected apt cache cleanup")
	}
}

func TestCleanCommands_Apk(t *testing.T) {
	opt := NewOptimizer("alpine:3.19")
	cmd := opt.CleanCommands("apk")
	if !strings.Contains(cmd, "--no-cache") {
		t.Error("expected --no-cache for apk")
	}
}
