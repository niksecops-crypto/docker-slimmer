package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/niksecops-crypto/docker-slimmer/pkg/optimizer"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "slimmer",
	Short:   "Optimize Docker images: analyze Dockerfiles and measure real image sizes",
	Version: version,
}

// ── generate ──────────────────────────────────────────────────────────────────

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an optimized multi-stage Dockerfile",
	RunE:  runGenerate,
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
	rootCmd.AddCommand(generateCmd)
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

// ── analyze ───────────────────────────────────────────────────────────────────

var analyzeCmd = &cobra.Command{
	Use:   "analyze <Dockerfile>",
	Short: "Analyze a Dockerfile and report improvement opportunities",
	Args:  cobra.ExactArgs(1),
	RunE:  runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
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
		fmt.Println("\n  No issues found.")
		return nil
	}

	fmt.Printf("\n  Issues (%d):\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("    %d. %s\n", i+1, issue)
	}
	return nil
}

// ── measure ───────────────────────────────────────────────────────────────────

var measureCmd = &cobra.Command{
	Use:   "measure [IMAGE] [--before IMAGE --after IMAGE]",
	Short: "Measure real image sizes via Docker daemon",
	Long: `Connects to the local Docker daemon and reports the actual uncompressed
size of one image, or computes the size reduction between two images.

Single image:
  slimmer measure myapp:latest

Before/after comparison:
  slimmer measure --before myapp:v1-legacy --after myapp:v2-optimized`,
	RunE: runMeasure,
}

var (
	measureBefore string
	measureAfter  string
)

func init() {
	measureCmd.Flags().StringVar(&measureBefore, "before", "", "Image reference to use as baseline")
	measureCmd.Flags().StringVar(&measureAfter, "after", "", "Image reference to compare against baseline")
	rootCmd.AddCommand(measureCmd)
}

func runMeasure(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	// Comparison mode: --before and --after
	if measureBefore != "" && measureAfter != "" {
		result, err := optimizer.CompareImages(ctx, measureBefore, measureAfter)
		if err != nil {
			return err
		}

		fmt.Printf("Image size comparison\n")
		fmt.Printf("  Before  %-40s  %s  (%d layers)\n",
			result.Before.Ref, result.Before.HumanSize(), result.Before.Layers)
		fmt.Printf("  After   %-40s  %s  (%d layers)\n",
			result.After.Ref, result.After.HumanSize(), result.After.Layers)
		fmt.Println()

		if result.SavedBytes >= 0 {
			fmt.Printf("  Saved   %s  (%.1f%% reduction)\n",
				optimizer.FormatBytes(result.SavedBytes), result.ReductionPct)
		} else {
			fmt.Printf("  Delta   +%s  (after image is larger)\n",
				optimizer.FormatBytes(-result.SavedBytes))
		}
		return nil
	}

	// Single image mode
	if len(args) == 0 {
		return fmt.Errorf("provide an image reference or use --before/--after for comparison")
	}

	stats, err := optimizer.InspectImage(ctx, args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Image:        %s\n", stats.Ref)
	fmt.Printf("ID:           %s\n", stats.ID)
	fmt.Printf("Size:         %s\n", stats.HumanSize())
	fmt.Printf("Layers:       %d\n", stats.Layers)
	fmt.Printf("Architecture: %s/%s\n", stats.OS, stats.Architecture)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
