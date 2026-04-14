package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/niksecops-crypto/docker-slimmer/pkg/optimizer"
	"github.com/spf13/cobra"
)

var (
	baseImage string
	artifacts []string
	output    string
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var rootCmd = &cobra.Command{
		Use:   "slimmer",
		Short: "Docker-Slimmer: Transform legacy Dockerfiles into optimized multi-stage builds",
		Long: `Docker-Slimmer is a production-ready utility for automating the optimization 
of Docker images. It analyzes current build patterns and generates 
highly-efficient multi-stage Dockerfiles based on security best practices (Distroless/Alpine).`,
		Run: func(cmd *cobra.Command, args []string) {
			slog.Info("Starting optimization process", "base", baseImage, "artifacts", artifacts)

			opt := optimizer.NewOptimizer(baseImage)
			optimizedContent := opt.Optimize(artifacts, map[string]string{"APP_ENV": "production"})

			if output != "" {
				err := os.WriteFile(output, []byte(optimizedContent), 0644)
				if err != nil {
					slog.Error("Failed to write output file", "error", err)
					os.Exit(1)
				}
				slog.Info("Optimized Dockerfile generated successfully", "path", output)
			} else {
				fmt.Println(optimizedContent)
			}
		},
	}

	rootCmd.Flags().StringVarP(&baseImage, "base", "b", "golang:1.22-alpine", "Base image for the build stage")
	rootCmd.Flags().StringSliceVarP(&artifacts, "artifacts", "a", []string{"/app/binary"}, "List of artifacts to copy to the runtime stage")
	rootCmd.Flags().StringVarP(&output, "output", "o", "Dockerfile.optimized", "Path to save the optimized Dockerfile")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
