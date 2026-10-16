# Buttprint API

Go HTTP service: receives location/time, calls Jackfruit, scores environmental inputs, and generates SVG plus metadata. Use go.mod and existing code for current versions and conventions.

- Preserve separation between HTTP handling, orchestration, Jackfruit access, scoring, and rendering. The renderer receives scores, not raw environmental values.
- Preserve per-variable coordinates/timestamps and nullable lineage; check the actual upstream contract before changing missing-data behavior.
- Follow existing interface seams, explicit error handling, request-context propagation, structured logging, and RFC 3339 timestamps with timezones. Keep scores in [0,1] and scoring weights normalized.
- Start with focused Go tests. Standard checks: go test ./..., go vet ./..., and make check. The server entry point is go run ./cmd/buttprint; do not start it merely to inspect the project.
