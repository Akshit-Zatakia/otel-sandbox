# otel-sandbox (Phase 1)

This is a minimal starter for the *Local OpenTelemetry Sandbox* CLI (lightweight mode using direct process management).

## What is included
- Cobra CLI with `up`, `down`, `status` commands.
- Manager that starts binaries found in `./bin` or in `PATH` and records PIDs in `state.json`.
- Minimal collector and prometheus configs in `assets/`.

## How to use
1. Build the binary:

```bash
cd cmd/otel-sandbox
go build -o ../../otel-sandbox
```

2. Place platform binaries in `./bin` or ensure they are available on PATH:
- `otelcol` (OpenTelemetry Collector binary)
- `jaeger-all-in-one` (Jaeger all-in-one binary)
- `prometheus` (Prometheus binary)

> For an easy start, you can download official builds and place them in `./bin`.

3. Start the sandbox:

```bash
./otel-sandbox up
./otel-sandbox status
# open Jaeger at http://localhost:16686
# open Prometheus at http://localhost:9090 

# verify
./otel-sandbox verify

# stop
./otel-sandbox down
```

## Next steps I can implement
- `download` command to auto-download and cache the required binaries for your platform
- Better cross-platform signal handling (Windows support)
- `verify` command to emit sample telemetry
- Lightweight embedded UI for unified view