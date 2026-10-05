# ✅ THE "AFTER" — multi-stage, minimal, non-root.
#
# Stage 1 compiles a static binary; stage 2 is a distroless image that contains
# *only* that binary — no shell, no package manager, no toolchain. The result is
# ~95% smaller and ships with effectively no OS-package CVEs.
#
# Tags are pinned to a version; for production, pin to a digest (@sha256:...) so
# the base can't change under you — that's covered in the hardening case study.

# ---- build stage ----
FROM golang:1.24.7-alpine AS build

WORKDIR /src

# Copy manifests first so `go mod download` is cached until deps actually change.
COPY app/go.mod ./
RUN go mod download

# Then the source. Changing code invalidates only from here down.
COPY app/ ./

ARG VERSION=dev
# CGO off + static link so the binary runs on a scratch/distroless base.
# -trimpath strips local paths; -s -w drop debug info to shrink the binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /orders-api .

# ---- runtime stage ----
# distroless "static:nonroot" = no shell, no libc, runs as UID 65532 by default.
FROM gcr.io/distroless/static-debian12:nonroot

# Copy only the compiled binary from the build stage.
COPY --from=build /orders-api /orders-api

# Non-root, explicit. (distroless:nonroot already defaults to this UID.)
USER 65532:65532

EXPOSE 8080

# No shell in the image, so health is checked by Kubernetes probes (see k8s/),
# not a Docker HEALTHCHECK.
ENTRYPOINT ["/orders-api"]
