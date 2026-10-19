# syntax=docker/dockerfile:1
# Backend image (plan §9.1): static Go binary on distroless, running as nonroot.
# All data lives in PostgreSQL; the image holds no content.

FROM golang:1.26.2-alpine AS build
WORKDIR /src
COPY . .
# Cache mounts: only the modules the binary needs are downloaded, not test deps.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=30s --retries=3 CMD ["/api", "healthcheck"]
ENTRYPOINT ["/api"]
CMD ["serve"]
