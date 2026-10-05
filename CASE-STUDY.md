# From manual deploys to a hardened pipeline — a 99% smaller image, zero-touch releases

**The work:** took a Go service that shipped by hand from a bloated, root-running
container, and rebuilt how it gets to production — a multi-stage distroless
image and a tested, vulnerability-scanned CI/CD pipeline.

---

## The problem

The service was deployed the way a lot of small teams deploy: by hand. A single
naive Dockerfile (`FROM golang:latest`, one stage, running as root), seven manual
steps to cut a release, no tests gating it, no image scan, and a rollback that
only existed in someone's memory. Every deploy was an event, and a bad build had
nothing standing between it and production.

Concretely, the baseline looked like this:

- **Image:** the full Go toolchain, module cache, and source all shipped inside
  the production image — roughly a gigabyte — running as **root**.
- **Release:** 7 manual commands from a laptop, no automated checks.
- **Safety:** no test gate, no vulnerability scan, `:latest` tags everywhere so
  you couldn't tell what was actually running.

## What I changed

1. **Multi-stage build.** One stage compiles a static, stripped binary; the
   final image is `distroless/static:nonroot` and contains *only* that binary —
   no shell, no package manager, no toolchain.
2. **Non-root by construction.** Runs as UID 65532. On Kubernetes it's locked
   down further: read-only root filesystem, all Linux capabilities dropped,
   `allowPrivilegeEscalation: false`, seccomp `RuntimeDefault`, and right-sized
   resource requests/limits.
3. **A real pipeline.** Every push runs `go vet` and the race-enabled test
   suite, then builds the image **once**, scans it with Trivy (the build fails
   on any HIGH/CRITICAL), and on `main` pushes to the registry and rolls out with
   `kubectl rollout status` gating success. Layer + module caching keeps repeat
   builds fast.
4. **Safe rollouts.** `maxUnavailable: 0`, liveness/readiness probes, and an
   immutable `sha-<commit>` tag on every image, so what's running is always
   traceable and a rollback is one `kubectl rollout undo`.

## The result

| Dimension | Before | After |
| --- | --- | --- |
| Image size | ~1.1 GB † | **~8 MB** † |
| HIGH/CRITICAL vulns (OS packages) | dozens † | **0** † |
| Runs as root | yes | **no** (UID 65532, read-only FS, caps dropped) |
| Dockerfile lint issues (hadolint) | 1 | **0** |
| Manual steps per release | 7 | **0** |
| Automated quality gates | 0 | **2** (tests + image scan) |
| Rollback | manual, ad-hoc | **one command**, immutable tags |

<sub>† Image size and vulnerability counts are produced by [`measure.sh`](./measure.sh)
on a machine with Docker; the binary is a measured 5.7 MB static build, and the
distroless base adds ~2 MB. Lint, step count, and gate figures are measured in
this repo. Re-run `./measure.sh` to regenerate the Docker-dependent numbers for
your own registry.</sub>

**Why it matters to the business:** deploys stopped being risky manual events.
A broken or vulnerable build can no longer reach production — it's caught by the
test gate or the scanner first — and the ~99% smaller image pulls and starts
faster, with a fraction of the attack surface.

---

*Stack: Go, Docker (multi-stage, distroless), GitHub Actions, Trivy, Kubernetes.
Repo includes the full before/after so the numbers are reproducible.*
