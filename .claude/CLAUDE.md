# CLAUDE.md

This file provides guidance to Claude Code when working in the Buttprint API repository.

## What This Is

Go HTTP service. The brains of Buttprint. Sits between the SvelteKit FE and the private Jackfruit data API:

```
Browser (FE) → Buttprint API (this, public) → Jackfruit (Go, private network)
```

Accepts `(lat, lon, timestamp)` from the FE (or resolves location from request IP), queries Jackfruit for raw environmental data, normalizes scores, generates a parametric SVG butt, returns SVG + metadata.


## Commands

```bash
go run ./cmd/buttprint              # Start server (default port 8080)
go build -o bin/buttprint ./cmd/buttprint  # Build binary
go test ./...                       # Run all tests
go vet ./...                        # Static analysis
```

Go 1.26+. Net/http standard library. No CGO.

## Architecture

### Module layout

```
cmd/
└── buttprint/
    └── main.go                 ← entry point, HTTP server setup

internal/
├── api/
│   ├── handler.go              ← HTTP handlers, ipResolver interface, route registration
│   ├── request.go              ← query parameter parsing and validation
│   ├── response.go             ← response structs and JSON serialization
│   └── client_ip.go            ← X-Forwarded-For extraction
├── config/
│   └── config.go               ← environment variable loading
├── domain/
│   ├── buttprint.go            ← service orchestrator, domain types, provider/scorer/renderer interfaces
│   └── errors.go               ← ErrNoData, ErrUpstream
├── geoloc/
│   ├── maxmind.go              ← MaxMind GeoLite2 implementation
│   └── errors.go               ← ErrPrivateIP, ErrLookupFailed
├── jackfruit/
│   └── client.go               ← HTTP client for Jackfruit API
├── scoring/
│   └── scorer.go               ← normalization + composite scoring
└── render/
    ├── svg.go                  ← parametric SVG generation (Bézier geometry, droplets)
    └── color.go                ← HSL color space, warmth ramp interpolation
```

### API contract

```
GET /buttprint?lat={float}&lon={float}&timestamp={RFC3339}
GET /health  → 204 No Content
```

All params optional — no coords triggers IP geolocation, no timestamp defaults to now.

### Jackfruit integration

```
GET jackfruit:8080/v1/environmental?lat=X&lon=Y&timestamp=T&variables=pm2p5,pm10,temperature,humidity,dewpoint
```

Called over private K8s network (ClusterIP). Pass Jackfruit's `variables` array through to the FE response. Lineage (`lineage` field per variable) is optional — Jackfruit returns null when Postgres lookup fails; treat as nullable throughout.

### Scoring

`Thiccness` is the composite score; the others express specific atmospheric qualities:

| Score | Primary inputs | Drives visually |
|---|---|---|
| `Thiccness` | temperature + humidity + pm2p5 + pm10 | Butt volume and Bézier curve geometry |
| `Sweatiness` | dewpoint | Surface sheen, highlight opacity |
| `Irritation` | pm2p5 (weighted higher) + pm10 | Blush redness, opacity |
| `Warmth` | temperature only | Gradient hue (cool blue → warm red) |

Missing variables are a fatal error — Jackfruit fails the entire request when a variable has no data, so partial input to the scorer indicates a programming bug. The renderer receives only `Score` — never raw variable values. All normalization lives in the scorer.

## Key Design Decisions

- **Interfaces for swappability:** `ipResolver` for IP geoloc (MVP: MaxMind GeoLite2, returns coords only), `Renderer` for SVG (MVP: parametric SVG, future: ASCII, diffusion). Design against the interface, not the implementation.
- **Lineage pass-through:** Don't flatten Jackfruit's per-variable response. Each variable carries its own `ref_timestamp`, `actual_lat/lon`, and `lineage` because variables may come from different datasets or grids.
- **Temperature unit conversion:** Jackfruit stores and returns temperature in °C. No conversion needed in Buttprint API — unit conversion happens at ingestion time in the pipeline.
- **Rate limiting:** In-memory per-IP (MVP). `golang.org/x/time/rate` token bucket. Behind an interface for future Redis upgrade.

## Go Conventions

- `internal/` for all non-exported packages
- Explicit error handling — no panics in request path
- Context propagation for cancellation
- `slog` with JSON handler for structured logging
- Environment variables for all configuration
- Standard library HTTP server with Go 1.22+ routing: `mux.HandleFunc("GET /path", handler)`
- Timestamps: `time.RFC3339` constant, `time.Parse` / `time.Format`

## Code Review

When reviewing PRs (automated or `@claude`-triggered):

- **Flag** unchecked error returns, missing context propagation, incorrect domain error mapping (`ErrNoData`/`ErrUpstream`), RFC 3339 violations, nil pointer risks on nullable lineage, broken interface contracts, response body leaks, score range violations (must be [0,1]), scoring weight drift (must sum to 1.0)
- **Skip** formatting and struct alignment (gofmt handles these), pre-existing debt unrelated to the PR
- **Verify** with `go vet ./...`, `go test ./...`, `make check` before posting findings
- **Severity tags:** 🔴 must-fix, 🟡 nit, 🟣 pre-existing
