package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	downloader "github.com/Akshit-Zatakia/otel-sandbox/internal/binaries"
	"github.com/Akshit-Zatakia/otel-sandbox/internal/verify"
)

// State keeps track of running processes' PIDs
type State struct {
	Processes map[string]int `json:"processes"`
}

type Manager struct {
	statePath string
	state     State
}

func NewManager(statePath string) *Manager {
	m := &Manager{statePath: statePath}
	m.loadState()
	return m
}

func (m *Manager) loadState() {
	m.state = State{Processes: map[string]int{}}
	f, err := os.Open(m.statePath)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewDecoder(f).Decode(&m.state)
}

func (m *Manager) saveState() error {
	f, err := os.Create(m.statePath)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(&m.state)
}

// Up starts the collector, jaeger and prometheus if available on PATH or ./bin
func (m *Manager) Up() error {
	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	fmt.Println("🚀 Starting OTel Sandbox...")

	// Download binaries if not present - uncomment this line
	if err := m.DownloadBinaries(); err != nil {
		fmt.Printf("⚠️  Warning: failed to download binaries: %v\n", err)
		fmt.Println("💡 You can manually place binaries in ./bin/ or ensure they're in your PATH")
	}

	fmt.Println("\n📦 Starting services...")

	// Start OpenTelemetry Collector (otelcol)
	if _, ok := m.state.Processes["otelcol"]; !ok {
		fmt.Print("  🔧 Starting OTel Collector... ")
		pid, err := m.startProcess("otelcol", []string{"--config", "assets/collector_config.yaml"}, "logs/otelcol.log")
		if err != nil {
			fmt.Printf("❌\n     Error: %v\n", err)
			fmt.Println("     💡 Check logs/otelcol.log for details")
		} else {
			m.state.Processes["otelcol"] = pid
			fmt.Printf("✅ (pid: %d)\n", pid)
		}
	} else {
		fmt.Printf("  ✅ OTel Collector already running (pid: %d)\n", m.state.Processes["otelcol"])
	}

	// Start Jaeger (jaeger-all-in-one)
	if _, ok := m.state.Processes["jaeger"]; !ok {
		fmt.Print("  🔧 Starting Jaeger... ")
		pid, err := m.startProcess("jaeger-all-in-one", []string{"--collector.zipkin.host-port=9411", "--collector.otlp.grpc.host-port=:14317", "--collector.otlp.http.host-port=:14318"}, "logs/jaeger.log")
		if err != nil {
			fmt.Printf("❌\n     Error: %v\n", err)
			fmt.Println("     💡 Check logs/jaeger.log for details")
		} else {
			m.state.Processes["jaeger"] = pid
			fmt.Printf("✅ (pid: %d)\n", pid)
		}
	} else {
		fmt.Printf("  ✅ Jaeger already running (pid: %d)\n", m.state.Processes["jaeger"])
	}

	// Start Prometheus if available
	if _, ok := m.state.Processes["prometheus"]; !ok {
		fmt.Print("  🔧 Starting Prometheus... ")
		pid, err := m.startProcess("prometheus", []string{"--config.file=assets/prometheus.yml", "--storage.tsdb.path=./prometheus-data"}, "logs/prometheus.log")
		if err != nil {
			fmt.Printf("❌\n     Error: %v\n", err)
			fmt.Println("     💡 Check logs/prometheus.log for details")
		} else {
			m.state.Processes["prometheus"] = pid
			fmt.Printf("✅ (pid: %d)\n", pid)
		}
	} else {
		fmt.Printf("  ✅ Prometheus already running (pid: %d)\n", m.state.Processes["prometheus"])
	}

	// Save state first
	if err := m.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Wait for services to initialize
	fmt.Println("\n⏳ Waiting for services to initialize...")
	time.Sleep(3 * time.Second)

	fmt.Println("\n🎉 OTel Sandbox is ready!")
	fmt.Println("\n📊 Available UIs:")
	if _, ok := m.state.Processes["jaeger"]; ok {
		fmt.Println("   🔍 Jaeger UI:     http://localhost:16686")
	}
	if _, ok := m.state.Processes["prometheus"]; ok {
		fmt.Println("   📈 Prometheus UI: http://localhost:9090")
	}

	fmt.Println("\n🧪 Next steps:")
	fmt.Println("   1. Run 'otel-sandbox verify' to test telemetry collection")
	fmt.Println("   2. Run 'otel-sandbox export --format summary' to view collected data")
	fmt.Println("   3. Check 'otel-sandbox status' to see process status")

	return nil
}

// Down stops all processes recorded in state
func (m *Manager) Down() error {
	var finalErr error
	for name, pid := range m.state.Processes {
		err := m.stopPID(pid)
		if err != nil {
			fmt.Printf("failed to stop %s (pid %d): %v\n", name, pid, err)
			finalErr = err
		} else {
			fmt.Printf("stopped %s (pid %d)\n", name, pid)
			delete(m.state.Processes, name)
		}
	}
	_ = m.saveState()
	return finalErr
}

// Status prints status of each tracked process
func (m *Manager) Status() error {
	if len(m.state.Processes) == 0 {
		fmt.Println("No running sandbox processes recorded.")
		return nil
	}
	for name, pid := range m.state.Processes {
		alive := m.isAlive(pid)
		status := "stopped"
		if alive {
			status = "running"
		}
		fmt.Printf("%s: pid=%d status=%s\n", name, pid, status)
	}
	return nil
}

func (m *Manager) DownloadBinaries() error {
	return downloader.DownloadAll("./bin")
}

func (m *Manager) Verify() error {
	return verify.Run(verify.VerifyConfig{
		CollectorEndpoint: "localhost:4317",
		Timeout:           5 * time.Second,
	})
}

// startProcess looks for executable in ./bin or PATH, starts it with args and redirects logs
func (m *Manager) startProcess(execName string, args []string, logfile string) (int, error) {
	path := m.findExecutable(execName)
	if path == "" {
		return 0, errors.New("executable not found: " + execName + " (place binary in ./bin or ensure in PATH)")
	}

	cmd := exec.Command(path, args...)
	// create log dir
	_ = os.MkdirAll(filepath.Dir(logfile), 0755)
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		cmd.Stdout = f
		cmd.Stderr = f
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	// start process in its own process group so we can kill children
	// On Unix use SysProcAttr
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	m.setProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	// detach: do not wait here. Save pid
	pid := cmd.Process.Pid
	// brief wait to see if process exits quickly
	time.Sleep(300 * time.Millisecond)
	if !m.isAlive(pid) {
		return pid, fmt.Errorf("process %s (pid %d) exited immediately, check %s", execName, pid, logfile)
	}

	return pid, nil
}

func (m *Manager) findExecutable(name string) string {
	// first check ./bin/name
	local := filepath.Join(".", "bin", name)
	if _, err := os.Stat(local); err == nil {
		return local
	}
	// fallback to PATH
	p, err := exec.LookPath(name)
	if err == nil {
		return p
	}
	return ""
}

func (m *Manager) stopPID(pid int) error {
	if !m.isAlive(pid) {
		return fmt.Errorf("pid %d not running", pid)
	}
	// kill process group
	// pgid, err := m.Getpgid(pid)
	// if err == nil {
	// 	// negative pid means kill process group
	// 	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	// 	// fallback
	// 	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	// 	return nil
	// }
	// // fallback: kill pid
	// _ = syscall.Kill(pid, syscall.SIGTERM)
	// _ = syscall.Kill(pid, syscall.SIGKILL)
	return m.killProcessGroup(pid)
}

func (m *Manager) isAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return m.isProcessAlive(pid)
}

// helper: read port from env or defaults (not used yet)
func intFromEnv(name string, def int) int {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
