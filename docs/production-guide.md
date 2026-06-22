# Docker-Slimmer: Production Guide

## Overview

Docker-Slimmer is a CLI tool with three commands:

| Command | What it does |
|---------|-------------|
| `analyze <Dockerfile>` | Scans an existing Dockerfile and reports optimization opportunities |
| `generate` | Generates a production-ready multi-stage Dockerfile from scratch |
| `measure` | Connects to Docker daemon and reports actual image sizes |

---

## Installation

### From source

```bash
git clone https://github.com/niksecops-crypto/docker-slimmer.git
cd docker-slimmer
go build -o slimmer ./cmd/slimmer
sudo mv slimmer /usr/local/bin/
```

### Docker (no Go toolchain required)

```bash
docker pull ghcr.io/niksecops-crypto/docker-slimmer:latest
alias slimmer='docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd):/workspace ghcr.io/niksecops-crypto/docker-slimmer:latest'
```

---

## Workflow: Legacy Image to Production-Grade

### Step 1 — Audit the existing Dockerfile

```bash
slimmer analyze ./Dockerfile
```

Example output for a common legacy image:

```
Dockerfile: ./Dockerfile
  Base image:    ubuntu:22.04
  Multi-stage:   false
  Distroless:    false
  Non-root user: false
  Stages:        1

  Issues (3):
    1. Single-stage build detected: consider multi-stage to separate build and runtime
    2. Base image is not distroless: runtime attack surface can be reduced significantly
    3. No explicit non-root user: add USER nobody or USER 65534
```

### Step 2 — Measure the baseline size

```bash
slimmer measure myapp:legacy
```

```
Image:        myapp:legacy
ID:           a1b2c3d4e5f6
Size:         387.12 MB
Layers:       18
Architecture: linux/amd64
```

### Step 3 — Generate an optimized Dockerfile

```bash
slimmer generate \
  --base golang:1.22-alpine \
  --artifacts /app/server,/app/config.yaml \
  --output Dockerfile.optimized
```

### Step 4 — Build and compare

```bash
docker build -f Dockerfile.optimized -t myapp:optimized .

slimmer measure --before myapp:legacy --after myapp:optimized
```

```
Image size comparison
  Before  myapp:legacy       387.12 MB  (18 layers)
  After   myapp:optimized     14.38 MB  (4 layers)

  Saved   372.74 MB  (96.3% reduction)
```

---

## Commercial Integration

### CI/CD pipeline (GitHub Actions)

```yaml
- name: Build and analyze
  run: |
    docker build -t $IMAGE_NAME:${{ github.sha }} .
    slimmer analyze ./Dockerfile
    slimmer measure $IMAGE_NAME:${{ github.sha }} >> $GITHUB_STEP_SUMMARY

- name: Compare with last release
  run: |
    slimmer measure \
      --before $IMAGE_NAME:latest \
      --after $IMAGE_NAME:${{ github.sha }} \
      >> $GITHUB_STEP_SUMMARY
```

### GitLab CI

```yaml
image-analysis:
  stage: test
  script:
    - slimmer analyze Dockerfile
    - slimmer measure --before $CI_REGISTRY_IMAGE:latest --after $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  artifacts:
    reports:
      dotenv: image_stats.env
```

---

## Base Image Selection Guide

| Application type | Recommended base | Why |
|-----------------|-----------------|-----|
| Go binary (CGO disabled) | `gcr.io/distroless/static-debian12` | No libc, smallest possible |
| Go binary (CGO enabled) | `gcr.io/distroless/base-debian12` | Includes glibc only |
| Python application | `gcr.io/distroless/python3-debian12` | No pip, no shell |
| Java application | `gcr.io/distroless/java21-debian12` | JRE only, no JDK |
| Node.js application | `gcr.io/distroless/nodejs20-debian12` | Node runtime only |

---

## Distroless Security Benefits

Distroless images eliminate the most common CVE sources:

- No shell (`/bin/sh`, `/bin/bash`) — prevents RCE exploitation via shell injection
- No package manager (`apt`, `apk`, `pip`) — no way to install additional packages post-deploy
- No OS utilities (`curl`, `wget`, `nc`) — removes common lateral movement tools
- Minimal libc — significantly smaller attack surface for memory corruption exploits

Use `trivy image` to verify the CVE count before and after:

```bash
trivy image myapp:legacy   --format table | grep CRITICAL
trivy image myapp:optimized --format table | grep CRITICAL
```

---

## Troubleshooting

**`slimmer measure` returns "connect to Docker daemon: ... permission denied"**
The Docker socket requires access. Run with `sudo` or add your user to the `docker` group:
```bash
sudo usermod -aG docker $USER
```

**Generated Dockerfile fails to build — binary not found**
Ensure the artifact path matches the actual build output. Check with:
```bash
docker run --rm myapp:builder ls -la /app/
```

**Image runs but crashes at startup**
Distroless has no shell — if your binary executes shell scripts or uses `exec`, it will fail. Ensure all dependencies (shared libraries, config files) are explicitly `COPY`'d.
