package optimizer

import (
	"fmt"
	"strings"
)

// Instruction represents a Dockerfile instruction
type Instruction struct {
	Command string
	Args    string
}

// Optimizer analyzes and optimizes Dockerfile instructions
type Optimizer struct {
	baseImage string
}

func NewOptimizer(baseImage string) *Optimizer {
	return &Optimizer{baseImage: baseImage}
}

// Optimize generates a multi-stage Dockerfile based on analyzed artifacts
func (o *Optimizer) Optimize(artifacts []string, envVars map[string]string) string {
	var builder strings.Builder

	// Build Stage
	builder.WriteString(fmt.Sprintf("# Build stage\nFROM %s AS builder\n", o.baseImage))
	builder.WriteString("WORKDIR /app\n")
	
	// Add environment variables
	for k, v := range envVars {
		builder.WriteString(fmt.Sprintf("ENV %s=%s\n", k, v))
	}

	// Runtime Stage (Distroless or Alpine for minimal footprint)
	builder.WriteString("\n# Runtime stage\nFROM gcr.io/distroless/static-debian12:latest\n")
	builder.WriteString("COPY --from=builder /etc/passwd /etc/passwd\n")
	builder.WriteString("USER nobody\n")
	builder.WriteString("WORKDIR /app\n")

	// Copy artifacts from builder
	for _, art := range artifacts {
		builder.WriteString(fmt.Sprintf("COPY --from=builder %s %s\n", art, art))
	}

	builder.WriteString("\nENTRYPOINT [\"/app/binary\"]\n")

	return builder.String()
}

// CleanCommands returns optimized shell commands (removing caches, temporary files)
func (o *Optimizer) CleanCommands(pkgManager string) string {
	switch pkgManager {
	case "apt":
		return "apt-get update && apt-get install -y --no-install-recommends %s && rm -rf /var/lib/apt/lists/*"
	case "apk":
		return "apk add --no-cache %s"
	default:
		return "echo 'Unknown package manager'"
	}
}
