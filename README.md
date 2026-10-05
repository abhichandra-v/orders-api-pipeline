# orders-api — CI/CD & container-hardening case study

A small Go HTTP service used to show a real before/after: a naive, root-running,
~1 GB container deployed by hand → a **~8 MB non-root distroless image** shipped
by a tested, vulnerability-scanned **CI/CD pipeline**.

👉 **Read the write-up: [CASE-STUDY.md](./CASE-STUDY.md)**

## What's here

| Path | What it is |
| --- | --- |
| `app/` | The Go service (health/readiness probes, graceful shutdown, tests) |
| `before/Dockerfile` | ❌ the naive, single-stage, root image (the baseline) |
| `before/deploy-notes.md` | how it was deployed by hand (7 manual steps) |
| `Dockerfile` | ✅ the multi-stage, distroless, non-root image |
| `.github/workflows/ci.yml` | test → build → Trivy scan → push → deploy |
| `k8s/` | hardened Deployment (non-root, read-only FS, caps dropped, probes) + Service |
| `measure.sh` | reproduces the image-size and vulnerability numbers |

## Run it locally

```bash
cd app
go test ./...
go run .          # serves on :8080
# in another shell:
curl localhost:8080/healthz
curl -X POST localhost:8080/orders -d '{"item":"widget","qty":3}'
curl localhost:8080/orders
```

## Reproduce the numbers

```bash
./measure.sh      # needs docker; trivy optional for vuln counts
```

## Build the hardened image

```bash
docker build -t orders-api:latest .
docker run --rm -p 8080:8080 orders-api:latest
```

---

Built by [Abhichandra](https://www.abhichandra.online) — SRE / Platform Engineer.
Available for scoped platform & container-security work.
