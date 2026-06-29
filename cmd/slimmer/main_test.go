package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/niksecops-crypto/docker-slimmer/pkg/optimizer"
)


type mockCLIRunner struct {
	runFunc func(ctx context.Context, name string, arg ...string) ([]byte, error)
}

func (m *mockCLIRunner) Run(ctx context.Context, name string, arg ...string) ([]byte, error) {
	return m.runFunc(ctx, name, arg...)
}

func captureOutput(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	f()

	wOut.Close()
	wErr.Close()

	var bufOut strings.Builder
	var bufErr strings.Builder

	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

func resetFlags() {
	baseImage = "golang:1.22-alpine"
	artifacts = []string{"/app/binary"}
	output = "Dockerfile.optimized"
	measureBefore = ""
	measureAfter = ""
}

func runCommand(args []string) (string, string, error) {
	resetFlags()
	rootCmd.SetArgs(args)

	var err error
	outStr, errStr := captureOutput(func() {
		err = rootCmd.Execute()
	})

	return outStr, errStr, err
}

func TestCLI_Help(t *testing.T) {
	out, _, err := runCommand([]string{"--help"})
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	if !strings.Contains(out, "slimmer [command]") {
		t.Errorf("expected help output, got: %s", out)
	}
}

func TestCLI_Analyze_MissingArgs(t *testing.T) {
	_, _, err := runCommand([]string{"analyze"})
	if err == nil {
		t.Fatal("expected error for missing analyze argument, got nil")
	}
}

func TestCLI_Analyze_Success(t *testing.T) {
	content := `FROM ubuntu:22.04
RUN apt-get update && apt-get install -y curl
`
	tmpfile, err := os.CreateTemp("", "Dockerfile.test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpfile.Close()

	out, _, err := runCommand([]string{"analyze", tmpfile.Name()})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	if !strings.Contains(out, "Base image:    ubuntu:22.04") {
		t.Errorf("unexpected output: %s", out)
	}
	if !strings.Contains(out, "Multi-stage:   false") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestCLI_Generate(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "Dockerfile.gen")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpfile.Close()
	os.Remove(tmpfile.Name()) // remove it so we can test generation

	defer os.Remove(tmpfile.Name())

	_, _, err = runCommand([]string{
		"generate",
		"--base", "golang:1.21-alpine",
		"--artifacts", "/bin/myapp",
		"--output", tmpfile.Name(),
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "FROM golang:1.21-alpine") {
		t.Errorf("generated Dockerfile missing base image, got: %s", content)
	}
	if !strings.Contains(content, "COPY --from=builder /bin/myapp") {
		t.Errorf("generated Dockerfile missing artifact copy, got: %s", content)
	}
}

func TestCLI_Measure_MissingArgs(t *testing.T) {
	_, _, err := runCommand([]string{"measure"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "provide an image reference") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_Measure_Single(t *testing.T) {
	oldRunner := optimizer.Runner
	defer func() { optimizer.Runner = oldRunner }()

	mockJSON := `[{"Id": "sha256:abc123def456789xyz", "Size": 500000, "Architecture": "amd64", "Os": "linux", "RootFS": {"Layers": ["l1"]}}]`

	optimizer.Runner = &mockCLIRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(mockJSON), nil
		},
	}

	out, _, err := runCommand([]string{"measure", "my-app:latest"})
	if err != nil {
		t.Fatalf("measure command failed: %v", err)
	}

	if !strings.Contains(out, "Image:        my-app:latest") {
		t.Errorf("unexpected output: %s", out)
	}
	if !strings.Contains(out, "ID:           abc123def456") {
		t.Errorf("unexpected output: %s", out)
	}
	if !strings.Contains(out, "Size:         488.28 KB") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestCLI_Measure_Compare(t *testing.T) {
	oldRunner := optimizer.Runner
	defer func() { optimizer.Runner = oldRunner }()

	optimizer.Runner = &mockCLIRunner{
		runFunc: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			img := arg[3]
			if img == "before-image" {
				return []byte(`[{"Id": "sha256:1", "Size": 1000, "Architecture": "amd64", "Os": "linux", "RootFS": {"Layers": []}}]`), nil
			}
			if img == "after-image" {
				return []byte(`[{"Id": "sha256:2", "Size": 200, "Architecture": "amd64", "Os": "linux", "RootFS": {"Layers": []}}]`), nil
			}
			return nil, fmt.Errorf("unknown image")
		},
	}

	out, _, err := runCommand([]string{"measure", "--before", "before-image", "--after", "after-image"})
	if err != nil {
		t.Fatalf("measure comparison failed: %v", err)
	}

	if !strings.Contains(out, "Image size comparison") {
		t.Errorf("unexpected output: %s", out)
	}
	if !strings.Contains(out, "Before  before-image") {
		t.Errorf("unexpected output: %s", out)
	}
	if !strings.Contains(out, "Saved   800 B  (80.0% reduction)") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestCLI_InvalidCommand(t *testing.T) {
	_, _, err := runCommand([]string{"nonexistent-command"})
	if err == nil {
		t.Fatal("expected error for nonexistent command, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected unknown command error, got: %v", err)
	}
}

func TestCLI_InvalidFlag(t *testing.T) {
	_, _, err := runCommand([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected error for unknown global flag, got nil")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("expected unknown flag error, got: %v", err)
	}
}

func TestCLI_Analyze_FileNotFound(t *testing.T) {
	_, _, err := runCommand([]string{"analyze", "/nonexistent/path/Dockerfile"})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "read config") {
		t.Errorf("expected file not found error, got: %v", err)
	}
}

func TestCLI_Generate_InvalidFlag(t *testing.T) {
	_, _, err := runCommand([]string{"generate", "--nonexistent-flag"})
	if err == nil {
		t.Fatal("expected error for unknown generate flag, got nil")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("expected unknown flag error, got: %v", err)
	}
}

func TestCLI_Measure_InvalidFlagsCombo(t *testing.T) {
	// If only --before is provided without --after
	_, _, err := runCommand([]string{"measure", "--before", "some-image"})
	if err == nil {
		t.Fatal("expected error for incomplete before/after flags, got nil")
	}
	if !strings.Contains(err.Error(), "provide an image reference or use --before/--after") {
		t.Errorf("unexpected error: %v", err)
	}
}

