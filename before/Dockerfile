# ❌ THE "BEFORE" — a naive, common Dockerfile. Do not copy this.
# Kept in the repo on purpose: it's the baseline the case study measures against.
#
# What's wrong with it:
#   1. `golang:latest` — a mutable, ~800 MB tag. Builds aren't reproducible, and
#      the whole Go toolchain + source ship to production inside the final image.
#   2. Single stage — the image contains the compiler, caches, git, and your
#      source code. Huge attack surface and huge size.
#   3. Runs as root — the default. A container escape starts as root.
#   4. No .dockerignore — .git, local caches, and secrets can leak into layers.
#   5. No pinned version, no healthcheck signal, no build flags.
FROM golang:latest

WORKDIR /app
COPY app/ ./
RUN go build -o orders-api .

EXPOSE 8080
CMD ["./orders-api"]
