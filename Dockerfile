# syntax=docker/dockerfile:1

FROM alpine:3.23 AS geolite2
RUN apk add --no-cache wget
RUN --mount=type=secret,id=maxmind_credentials \
    CREDS=$(cat /run/secrets/maxmind_credentials) && \
    wget --user="${CREDS%%:*}" --password="${CREDS#*:}" \
      -qO /tmp/geolite2.tar.gz \
      "https://download.maxmind.com/geoip/databases/GeoLite2-City/download?suffix=tar.gz" && \
    tar -xzf /tmp/geolite2.tar.gz -C /tmp && \
    mkdir -p /data && \
    mv /tmp/GeoLite2-City_*/GeoLite2-City.mmdb /data/GeoLite2-City.mmdb

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/buttprint ./cmd/buttprint

FROM alpine:3.23
COPY --from=geolite2 /data/GeoLite2-City.mmdb /data/GeoLite2-City.mmdb
COPY --from=builder /bin/buttprint /bin/buttprint
EXPOSE 8080
ENTRYPOINT ["/bin/buttprint"]
