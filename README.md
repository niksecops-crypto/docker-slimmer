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

Legacy Docker images often carry 500MB+ of overhead (compilers, build tools, package caches). Docker-Slimmer can reduce image size by up to 90% while significantly improving security by removing shell access and common vulnerabilities (CVEs) found in base OS distributions.

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

Optimize a legacy Go-based Dockerfile by extracting a binary to a Distroless runtime:

```bash
./slimmer --base golang:1.22-alpine --artifacts /app/main --output Dockerfile.optimized
```

## Production Best Practices

- **Scan Before & After**: Use tools like `trivy` or `grype` to verify the reduction in vulnerabilities.
- **Test Entrypoints**: Ensure all runtime dependencies are explicitly copied to the final stage.
- **Distroless for Security**: Prefer `gcr.io/distroless/static` for statically linked binaries (Go, Rust).

## License

Distributed under the MIT License. See `LICENSE` for more information.

---
*Maintained by [niksecops-crypto](https://github.com/niksecops-crypto)*
