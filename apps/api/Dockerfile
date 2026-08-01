# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine3.24@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/belcanto-api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/belcanto-bootstrap-owner ./cmd/bootstrap-owner && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/belcanto-bootstrap-staff ./cmd/bootstrap-staff && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/belcanto-reissue-bootstrap-invitation ./cmd/reissue-bootstrap-invitation && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/belcanto-migrate ./cmd/migrate

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 10001 belcanto && \
    adduser -S -D -H -u 10001 -G belcanto belcanto
WORKDIR /app
COPY --from=build /out/belcanto-api /app/belcanto-api
COPY --from=build /out/belcanto-bootstrap-owner /app/belcanto-bootstrap-owner
COPY --from=build /out/belcanto-bootstrap-staff /app/belcanto-bootstrap-staff
COPY --from=build /out/belcanto-reissue-bootstrap-invitation /app/belcanto-reissue-bootstrap-invitation
COPY --from=build /out/belcanto-migrate /app/belcanto-migrate
USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O - http://127.0.0.1:8080/readyz >/dev/null || exit 1
ENTRYPOINT ["/app/belcanto-api"]
