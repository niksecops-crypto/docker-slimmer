# Changelog

## [1.1.0] - 2024-12-10
### Added
- `analyze` subcommand: parses existing Dockerfiles and reports issues
- Dockerfile parser (`pkg/optimizer/parser.go`) with multi-stage detection
- Distroless and non-root user detection in analysis
- Unit tests for optimizer and parser (~44% overall module statement coverage)
- GitHub Actions CI with cross-platform release builds
- Makefile with `build`, `test`, `lint` targets
- Dockerfile (distroless, multi-stage)
- `--version` flag

### Changed
- CLI restructured to subcommands: `slimmer generate` / `slimmer analyze`

## [1.0.0] - 2024-10-20
### Added
- Initial release: multi-stage Dockerfile generation
- Distroless runtime base selection
- apt/apk package manager cache cleanup helpers
