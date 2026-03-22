# Buttprint API

Go HTTP service that turns atmospheric conditions at a given place and time into a parametric SVG butt. Queries a private environmental data API ([Jackfruit](https://github.com/kacper-wojtaszczyk/jackfruit)), normalizes atmospheric variables into four scores — thiccness, sweatiness, irritation, warmth — and generates an SVG whose geometry and colour reflect the data.

## Examples

Atmospheric conditions shape the butt's size, colour, sheen, and blush. A few scenarios from real-world weather patterns:

<table>
<tr>
<td align="center"><img src="docs/examples/siberian_winter.svg" width="120"><br><sub>Siberian Winter<br>−20°C · 25% RH</sub></td>
<td align="center"><img src="docs/examples/temperate_spring.svg" width="120"><br><sub>Temperate Spring<br>15°C · 60% RH</sub></td>
<td align="center"><img src="docs/examples/humid_tropics.svg" width="120"><br><sub>Humid Tropics<br>32°C · 92% RH</sub></td>
<td align="center"><img src="docs/examples/delhi_haze.svg" width="120"><br><sub>Delhi Haze<br>35°C · PM2.5 200</sub></td>
<td align="center"><img src="docs/examples/wildfire_smoke.svg" width="120"><br><sub>Wildfire Smoke<br>25°C · PM2.5 200</sub></td>
</tr>
</table>

## How It Works

```
Request (lat, lon, timestamp)
  → Query Jackfruit for environmental data
  → Normalize variables into [0,1] scores
  → Scores drive Bézier curves, colour, opacity
  → JSON response with SVG + metadata
```

See [`docs/scoring-formula.md`](docs/scoring-formula.md) for scoring calibration and weights.

## Project Structure

```
cmd/buttprint/          Entry point
internal/
  api/                  HTTP handlers + request/response
  config/               Environment configuration
  domain/               Core types and interfaces
  jackfruit/            Upstream data API client
  scoring/              Normalization + composite scoring
  render/               Parametric SVG generation
docs/
  scoring-formula.md    Scoring calibration reference
  examples/             Generated SVG samples
```

## Running

Requires a running [Jackfruit](https://github.com/kacper-wojtaszczyk/jackfruit) instance as the upstream data source — see its README for local setup.

```bash
cp .env.example .env          # Configure environment (set JACKFRUIT_URL)
go run ./cmd/buttprint        # Start server on port 8080
```

## Testing

```bash
make test                     # All tests
make test-short               # Unit tests only
```

## API

```
GET /health → 204 No Content
GET /buttprint?lat=52.52&lon=13.40&timestamp=2026-03-08T14:00:00Z
```

## Related Repos

Part of the [Climacterium](https://github.com/kacper-wojtaszczyk?tab=repositories) ecosystem:

| Repo | Description |
|------|-------------|
| [jackfruit](https://github.com/kacper-wojtaszczyk/jackfruit) | Environmental data ingestion + serving (Go, Python, ClickHouse) |
| [buttprint-fe](https://github.com/kacper-wojtaszczyk/buttprint-fe) | Display layer (SvelteKit) |
| [climacterium-infra](https://github.com/kacper-wojtaszczyk/climacterium-infra) | Terraform + Kubernetes deployment (Scaleway) |
