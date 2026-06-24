# Contributing to docker-slimmer

## Getting Started

```bash
git clone https://github.com/niksecops-crypto/docker-slimmer.git
cd docker-slimmer
go mod download
make test
```

## Development Workflow

1. Fork the repo and create a branch: `git checkout -b feat/my-feature`
2. Make changes, add tests
3. `make test` — must pass
4. `make lint` — no errors
5. Open a PR against `main`

## Running Tests

```bash
make test    # full suite with race detector
```

No Docker daemon needed — the optimizer and parser are pure Go.

## Adding Optimizer Rules

New optimization rules go in `pkg/optimizer/optimizer.go`. Add a corresponding test case in `pkg/optimizer/optimizer_test.go`.

New Dockerfile analysis checks go in `pkg/optimizer/parser.go` → `AnalyzeIssues()`.

## Reporting Issues

Open a [GitHub Issue](https://github.com/niksecops-crypto/docker-slimmer/issues) with:
- Input Dockerfile (or minimal reproduction)
- Expected vs actual output
- Tool version (`slimmer --version`)
