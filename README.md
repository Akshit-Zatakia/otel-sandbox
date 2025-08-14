# otel-sandbox

A lightweight Local OpenTelemetry Sandbox CLI tool for developers to quickly set up, test, and experiment with OpenTelemetry instrumentation locally without Docker.

## Features

🚀 **Quick Setup**: Start OTel Collector, Jaeger, and Prometheus with a single command  
🔍 **Smart Detection**: Automatically analyzes your project and suggests appropriate instrumentation  
✅ **Easy Verification**: Test your telemetry setup with sample data  
📊 **Data Export**: View collected telemetry data in multiple formats  
⬇️ **Auto Download**: Automatically downloads required binaries for your platform  

## What's Included

- **OpenTelemetry Collector**: Receives and processes telemetry data
- **Jaeger**: Distributed tracing UI at http://localhost:16686
- **Prometheus**: Metrics storage and UI at http://localhost:9090
- **Project Analysis**: Detects Go, Node.js, Python, and Java projects
- **Framework Detection**: Supports Gin, Express, and other popular frameworks
- **Multiple Export Formats**: JSON, CSV, and summary formats

## Quick Start

1. **Build the binary:**
```bash
go build -o otel-sandbox cmd/main.go
```

2. **Download required binaries (optional - they can also be in your PATH):**
```bash
./otel-sandbox download
```

3. **Start the sandbox:**
```bash
./otel-sandbox up
```

4. **Verify your setup:**
```bash
./otel-sandbox verify
```

5. **View collected data:**
```bash
./otel-sandbox export --format summary
```

6. **Stop the sandbox:**
```bash
./otel-sandbox down
```

### Commands
**Core Commands**
- `up` - Start all services (OTel Collector, Jaeger, Prometheus)
- `down` - Stop all services and clean up
- `status` - Check status of running services

**Analysis & Verification**
- `verify` - Send sample telemetry data to test your setup
- `export` - Export collected telemetry data in various formats

**Setup & Maintenance**
- `download` - Download required binaries for Collector, Jaeger, and Prometheus

### Command Examples
**Basic Usage**
```bash
# Start the sandbox
./otel-sandbox up

# Check what's running
./otel-sandbox status

# Stop everything
./otel-sandbox down
```

**Verification & Export**
```bash
# Verify your setup
./otel-sandbox verify

# Export collected data
./otel-sandbox export --format summary

# Export data as CSV
./otel-sandbox export --format csv --output telemetry.csv

# Export specific service data
./otel-sandbox export --service my-service --format json
```

**Binary Management**
```bash
# Download required binaries (optional - they can also be in your PATH)
./otel-sandbox download
```

### Troubleshooting
**Services won't start:**

- Check if ports 4317, 4318, 16686, 9090, 8889 are available
- Run ./otel-sandbox download to ensure binaries are present

**No telemetry data:**

- Run ./otel-sandbox verify to test the setup
- Check collector logs in the terminal output

### Contributing
- Fork the repository
- Create a feature branch
- Add tests for new functionality
- Run the test suite: go test ./...
- Submit a pull request
