# The "before" deploy — done by hand

This is how the service shipped before the pipeline existed. Every release was
a sequence of manual commands run from a laptop, with nothing stopping a bad
build from reaching production.

1. `go build -o orders-api .`
2. `docker build -f before/Dockerfile -t orders-api .`
3. `docker tag orders-api registry.example.com/orders-api:latest`
4. `docker push registry.example.com/orders-api:latest`
5. SSH to the box / run `kubectl set image …` by hand
6. Manually watch logs to see if it came up
7. If it broke: manually rebuild the previous version and push again (no
   rehearsed rollback, no record of what the previous good version was)

**Problems:** 7 manual steps, no tests gating the release, no vulnerability
scan, `:latest` everywhere (so you can't tell what's actually running), and a
rollback that only exists in someone's head. This is the baseline the pipeline
replaces.
