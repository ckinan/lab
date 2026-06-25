# lab

Personal Software Incubator.

An experimental monorepo and incubator for side projects, CLI utilities, and production services.

## Philosophy

**Zero-Friction Prototyping**: New app ideas start right here in `apps/`.

## Incubated Apps

### `gitz`
Fast local Git repository scanner and status summary utility.

### `cktop`
Terminal UI for real-time Linux system monitoring.
```bash
go install github.com/ckinan/lab/cmd/cktop@latest
```

### `ckstat`
One-shot CLI that prints current memory usage.
```bash
go install github.com/ckinan/lab/cmd/ckstat@latest

ckstat          # human-readable
ckstat -o json  # JSON output
```

### `ckagent`
Metrics agent that exposes `/proc` system data in Prometheus exposition format.
```bash
go install github.com/ckinan/lab/cmd/ckagent@latest

ckagent              # listens on :9110
ckagent -addr :9111  # custom port
```
*Metrics exposed at `http://localhost:9110/metrics`.*

**Exposed Metrics:**
- CPU, memory, swap, load averages
- Per-core CPU usage
- Disk I/O (reads, writes, io time)
- Network (bytes, packets, drops, errors per interface)
- Open file descriptors
- Socket counts (tcp, udp, raw, orphan, timewait)
- `vmstat` (page faults, paging, swapping)
- Process/task counts & system uptime
- `apt` last update and upgrade timestamps
- PSI (Pressure Stall Information: cpu, memory, io) system-wide and per-service via cgroups

---

## Simulation & Load Testing

### `ckapi`
Programmable HTTP target behavior simulator (delay, CPU burn, memory pressure, failure injection). Designed to drive synthetic metrics into Grafana/VictoriaMetrics.
```bash
cd ckapi && go install ./cmd

ckapi              # listens on :9120
ckapi -addr :9121  # custom port
```

#### Endpoints
- `POST /work` : Execute a behavior
- `POST /control` : Set default behavior for subsequent `/work` requests
- `GET /control` : Read current defaults
- `GET /metrics` : Prometheus exposition (`ckapi_*` + `go_memstats_*`)
- `GET /health` : Liveness check

#### Request Fields
- `delay_ms` *(int)* : Sleep N ms before responding
- `cpu_burn_ms` *(int)* : Spin CPU for N ms
- `mem_use_bytes` *(int)* : Allocate N bytes
- `mem_hold` *(bool)* : Hold allocation for request duration
- `fail` *(bool)* : Return 500
- `status_code` *(int)* : Return specific HTTP status code

#### Examples
```bash
# Set defaults: 50ms delay, no failure
curl -s -X POST localhost:9120/control \
  -H 'Content-Type: application/json' \
  -d '{"delay_ms":50,"fail":false}'

# Override single request: 800ms delay
curl -s -X POST localhost:9120/work \
  -H 'Content-Type: application/json' \
  -d '{"delay_ms":800}'
```

### `k6` Synthetic Load Generation
Continuous synthetic load testing scripts against `ckapi`.
```bash
k6 run k6/ckapi.js                                    # run locally
CKAPI_URL=http://host:9120/work k6 run k6/ckapi.js    # custom target
```

**Scenarios Included (`k6/ckapi.js`):**
- `steady` : 30 RPS constant, 20ms delay
- `latency-spike` : 20 RPS base, spikes to 150 RPS with 800ms delay every 5m
- `sawtooth-load` : 5–100 RPS sawtooth, 50ms delay
- `error-injection` : 10 RPS constant, `fail=true`
- `memory-hold` : 5 RPS constant, 10MB held per request

---

## Development

Requires Go 1.21+ and [Task](https://taskfile.dev).

```bash
task build          # Build all binaries
task test           # Run unit tests across all incubated packages
task run:agent      # Run ckagent locally
task run:api        # Run ckapi locally
task run:load       # Run k6 load generation locally
task install:agent  # Build & deploy ckagent to /usr/local/bin + restart systemd
task install:api    # Build & deploy ckapi to /usr/local/bin + restart systemd
task install:load   # Deploy k6 script + restart systemd load service
```
