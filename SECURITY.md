# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.x     | ✅        |
| < 1.0   | ❌        |

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report privately via GitHub Security Advisories:
👉 [Report a vulnerability](https://github.com/niksecops-crypto/docker-slimmer/security/advisories/new)

Or email: **security@niksecops.dev**

You will receive an acknowledgement within **48 hours** and a resolution timeline within **7 days**.

## Security Considerations

- docker-slimmer reads local Dockerfile paths — never pass untrusted paths from external input
- Generated Dockerfiles use `gcr.io/distroless` and `USER nobody` by default
- Always scan generated images with `trivy image` before deploying to production
