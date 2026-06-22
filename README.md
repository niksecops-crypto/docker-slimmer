# Docker-Slimmer: Production-Grade Image Optimizer

[![CI](https://github.com/niksecops-crypto/docker-slimmer/actions/workflows/ci.yml/badge.svg)](https://github.com/niksecops-crypto/docker-slimmer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/niksecops-crypto/docker-slimmer)](https://goreportcard.com/report/github.com/niksecops-crypto/docker-slimmer)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/GHCR-ghcr.io-blue?logo=docker)](https://github.com/niksecops-crypto/docker-slimmer/pkgs/container/docker-slimmer)

Docker-Slimmer is an automated tool designed to transform legacy, bloated Dockerfiles into highly-efficient, secure, and lightweight multi-stage builds. It helps DevOps engineers modernize inherited container images by applying industry best practices such as Distroless base images and aggressive cache cleaning.

## Key Features

- **Multi-Stage Build Automation**: Automatically splits your build process into build-time and runtime stages.
- **Distroless Runtime Support**: Uses Google's Distroless static images for minimal attack surface and smallest possible footprint.
- **Automated Cache Cleaning**: Integrated support for `apt` and `apk` package managers to remove temporary build files.
- **Security-First Approach**: Runs as `nobody` user by default and removes unnecessary shell binaries.
- **Artifact Isolation**: Only copies required binaries and configuration files to the final image.

## Why Use Docker-Slimmer?

Legacy Docker images often carry 500MB+ of overhead (compilers, build tools, package caches). Multi-stage builds with Distroless runtime images typically reduce image size by 60–90%, depending on the application. Use `slimmer measure` to get the actual numbers for your specific images rather than relying on estimates.

## Getting Started

### Prerequisites

- Go 1.22+ (to build the tool)
- Docker (to build the optimized images)

### Installation

```bash
git clone https://github.com/niksecops-crypto/docker-slimmer.git
cd docker-slimmer
go build -o slimmer ./cmd/slimmer
```

### Usage

**Generate** an optimized multi-stage Dockerfile:

```bash
./slimmer generate --base golang:1.22-alpine --artifacts /app/main --output Dockerfile.optimized
```

**Analyze** an existing Dockerfile for improvement opportunities:

```bash
./slimmer analyze ./Dockerfile
# Dockerfile: ./Dockerfile
#   Base image:    ubuntu:22.04
#   Multi-stage:   false
#   Distroless:    false
#   Non-root user: false
#
#   Issues (3):
#     1. Single-stage build detected: consider multi-stage to separate build and runtime
#     2. Base image is not distroless: runtime attack surface can be reduced significantly
#     3. No explicit non-root user: add USER nobody or USER 65534
```

**Measure** real image sizes via Docker daemon:

```bash
# Inspect a single image
./slimmer measure myapp:latest
# Image:        myapp:latest
# ID:           a1b2c3d4e5f6
# Size:         312.45 MB
# Layers:       14
# Architecture: linux/amd64

# Compare before/after optimization
./slimmer measure --before myapp:legacy --after myapp:optimized
# Image size comparison
#   Before  myapp:legacy       312.45 MB  (14 layers)
#   After   myapp:optimized     18.32 MB  (4 layers)
#
#   Saved   294.13 MB  (94.1% reduction)
```

## Production Best Practices

- **Measure, don't guess**: always run `slimmer measure --before ... --after ...` to confirm actual savings.
- **Scan for CVEs**: use `trivy image myapp:optimized` to verify the vulnerability reduction alongside size.
- **Distroless for security**: prefer `gcr.io/distroless/static` for statically linked binaries (Go, Rust).
- **Test entrypoints**: distroless images have no shell — ensure all runtime dependencies are explicitly copied.

## Documentation

- [Production Guide](docs/production-guide.md) — CI/CD integration, base image selection, distroless security benefits, troubleshooting

## License

Distributed under the MIT License. See `LICENSE` for more information.

---
*Maintained by [niksecops-crypto](https://github.com/niksecops-crypto)*
