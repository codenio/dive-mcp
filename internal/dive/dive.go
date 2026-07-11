package dive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// validSources are the --source values dive understands.
var validSources = map[string]bool{
	"":               true, // default (docker)
	"docker":         true,
	"podman":         true,
	"docker-archive": true,
}

// DefaultTimeout is used when DIVE_MCP_TIMEOUT is unset or unparsable.
const DefaultTimeout = 5 * time.Minute

// Runner locates and invokes the dive CLI, caching parsed analyses for the
// lifetime of the process so that multiple MCP tool calls against the same
// (image, source) pair only shell out to dive once.
type Runner struct {
	// BinPath is the resolved path to the dive executable.
	BinPath string
	// Timeout bounds each invocation of dive.
	Timeout time.Duration

	mu    sync.Mutex
	cache map[string]*Analysis
}

// NewRunner locates the dive binary (via DIVE_MCP_DIVE_PATH or PATH) and
// reads DIVE_MCP_TIMEOUT for the per-invocation timeout. It returns an error
// with actionable guidance if dive cannot be found.
func NewRunner() (*Runner, error) {
	bin, err := findDiveBinary()
	if err != nil {
		return nil, err
	}
	return &Runner{
		BinPath: bin,
		Timeout: timeoutFromEnv(),
		cache:   make(map[string]*Analysis),
	}, nil
}

func findDiveBinary() (string, error) {
	if p := os.Getenv("DIVE_MCP_DIVE_PATH"); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("dive binary not found at DIVE_MCP_DIVE_PATH=%q: %w", p, err)
		}
		return p, nil
	}
	p, err := exec.LookPath("dive")
	if err != nil {
		return "", fmt.Errorf(
			"dive binary not found on PATH: install it from https://github.com/wagoodman/dive " +
				"or set DIVE_MCP_DIVE_PATH to its location",
		)
	}
	return p, nil
}

func timeoutFromEnv() time.Duration {
	v := os.Getenv("DIVE_MCP_TIMEOUT")
	if v == "" {
		return DefaultTimeout
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	return DefaultTimeout
}

// Analyze runs (or fetches from cache) the dive analysis for image using the
// given source ("", "docker", "podman", or "docker-archive").
func (r *Runner) Analyze(ctx context.Context, image, source string) (*Analysis, error) {
	if image == "" {
		return nil, fmt.Errorf("image must not be empty")
	}
	if !validSources[source] {
		return nil, fmt.Errorf("invalid source %q: must be one of docker, podman, docker-archive", source)
	}

	key := source + "\x00" + image

	r.mu.Lock()
	if a, ok := r.cache[key]; ok {
		r.mu.Unlock()
		return a, nil
	}
	r.mu.Unlock()

	analysis, err := r.run(ctx, image, source)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cache[key] = analysis
	r.mu.Unlock()

	return analysis, nil
}

func (r *Runner) run(ctx context.Context, image, source string) (*Analysis, error) {
	tmp, err := os.CreateTemp("", "dive-mcp-*.json")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	args := []string{image, "--json", tmpPath}
	if source != "" {
		args = append(args, "--source", source)
	}

	cmd := exec.CommandContext(ctx, r.BinPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "CI=") // ensure --ci path is not implicitly taken

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("dive timed out after %s analyzing %q: %s", r.Timeout, image, stderr.String())
		}
		return nil, fmt.Errorf("dive failed analyzing %q: %w: %s", image, err, stderr.String())
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("reading dive output: %w", err)
	}

	var analysis Analysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("parsing dive JSON output: %w", err)
	}

	return &analysis, nil
}
