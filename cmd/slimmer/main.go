package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/niksecops-crypto/docker-slimmer/pkg/optimizer"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "slimmer",
	Short:   "Optimize Docker images: multi-stage builds, distroless bases, minimal attack surface",
	Version: version,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an optimized Dockerfile from scratch",
	RunE:  runGenerate,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze <Dockerfile>",
	Short: "Analyze an existing Dockerfile and report improvement opportunities",
	Args:  cobra.ExactArgs(1),
	RunE:  runAnalyze,
}

var (
	baseImage string
	artifacts []string
	output    string
)

func init() {
	generateCmd.Flags().StringVarP(&baseImage, "base", "b", "golang:1.22-alpine", "Builder base image")
	generateCmd.Flags().StringSliceVarP(&artifacts, "artifacts", "a", []string{"/app/binary"}, "Artifacts to copy to runtime stage")
	generateCmd.Flags().StringVarP(&output, "output", "o", "Dockerfile.optimized", "Output file path")

	rootCmd.AddCommand(generateCmd, analyzeCmd)
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("generating optimized Dockerfile", "base", baseImage, "artifacts", artifacts)

	opt := optimizer.NewOptimizer(baseImage)
	content := opt.Optimize(artifacts, map[string]string{"APP_ENV": "production"})

	if err := os.WriteFile(output, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	slog.Info("done", "path", output)
	return nil
}

func runAnalyze(_ *cobra.Command, args []string) error {
	path := args[0]

	df, err := optimizer.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parse %q: %w", path, err)
	}

	fmt.Printf("Dockerfile: %s\n", path)
	fmt.Printf("  Base image:    %s\n", df.BaseImage)
	fmt.Printf("  Multi-stage:   %v\n", df.IsMultiStage)
	fmt.Printf("  Distroless:    %v\n", df.HasDistroless)
	fmt.Printf("  Non-root user: %v\n", df.HasNobodyUser)
	fmt.Printf("  Stages:        %d\n", len(df.Stages))

	issues := optimizer.AnalyzeIssues(df)
	if len(issues) == 0 {
		fmt.Println("\n✓ No issues found")
		return nil
	}

	fmt.Printf("\nIssues (%d):\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("  %d. %s\n", i+1, issue)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
