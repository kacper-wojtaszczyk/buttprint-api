# Scoring Formula

The scorer converts raw environmental data into four normalized [0, 1] scores that drive the visual attributes of the parametric SVG butt.

## Calibration Ranges

Each raw variable is normalized to [0, 1] using a linear clamp against a calibration range. Values outside the range are clamped — never extrapolated.

The anchors: the flattest butt is a Siberian winter (cold, dry, clean); the thiccest is a Delhi summer (extreme heat, humid, heavy PM).

| Variable | Min (→ 0.0) | Max (→ 1.0) | Rationale |
|---|---|---|---|
| temperature | −30 °C | 48 °C | Siberian winter / Delhi extreme summer |
| humidity | 10 % | 98 % | Rarified desert / near-saturation |
| dewpoint | −20 °C | 32 °C | Siberian dry cold / extreme tropical monsoon |
| pm2p5 | 0 µg/m³ | 300 µg/m³ | Clean air / severely polluted (WHO 24h guideline: 15 µg/m³) |
| pm10 | 0 µg/m³ | 500 µg/m³ | Clean air / severe dust or pollution event |

## Scores

### Thickness (composite)

Weighted sum of four normalized inputs. Represents overall atmospheric oppressiveness — higher means heavier, more suffocating air. Drives butt volume and Bezier curve geometry in the renderer.

```
Thickness = 0.30 * norm(temperature)
          + 0.30 * norm(humidity)
          + 0.25 * norm(pm2p5)
          + 0.15 * norm(pm10)
```

Temperature and humidity are weighted equally and together dominate the score (~60%) because perceived oppressiveness is primarily a heat-moisture phenomenon. PM2.5 gets more weight than PM10 because fine particulates are more acutely harmful and contribute more to the "thick air" sensation.

### Sweatiness

Direct normalization of dewpoint. Dew point is the most physically accurate single predictor of sweaty discomfort — meteorologists use it precisely for this purpose (above ~16 °C feels humid, above 21 °C feels oppressive). Drives highlight opacity and surface sheen.

```
Sweatiness = norm(dewpoint)
```

### Irritation

Weighted sum of normalized particulate matter values. Represents airway irritation from inhaled particles. Drives blush redness and opacity.

```
Irritation = 0.65 * norm(pm2p5)
           + 0.35 * norm(pm10)
```

PM2.5 gets higher weight because fine particles penetrate deeper into the lungs and are more acutely harmful.

### Warmth

Direct normalization of temperature. Not a comfort score — a colour signal for the renderer to map onto a cool-blue → warm-red gradient.

```
Warmth = norm(temperature)
```
