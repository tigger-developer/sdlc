//go:build unix

// ABOUTME: Runs one tool-free Claude safe-mode experiment with bounded runtime.
// ABOUTME: Keeps experimental session state separate from installed SDLC audit gates.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed help.md
var help string

//go:embed config.yaml
var config []byte

//go:embed response.schema.json
var schema string

type response struct {
	Verdict  string   `json:"verdict" yaml:"verdict"`
	Findings []string `json:"findings" yaml:"findings"`
}

type result struct {
	SessionID string   `yaml:"session_id"`
	Response  response `yaml:"response"`
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	flags := flag.NewFlagSet("claude-audit", flag.ContinueOnError)
	flags.Usage = func() { fmt.Fprint(os.Stdout, help) }
	state := flags.String("state-dir", "", "scratch state directory")
	resume := flags.String("resume", "", "native session ID")
	dryRun := flags.Bool("dry-run", false, "show arguments without invoking")
	version := flags.Bool("version", false, "show prototype version")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *version {
		fmt.Println("claude-audit prototype W007")
		return 0
	}
	if *state == "" || flags.NArg() != 0 {
		flags.Usage()
		return 2
	}
	if err := execute(*state, *resume, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, "claude-audit:", err)
		return 1
	}
	return 0
}

func execute(state, resume string, dryRun bool) error {
	state, err := filepath.Abs(state)
	if err != nil {
		return err
	}
	info, err := os.Stat(state)
	if err != nil {
		return fmt.Errorf("scratch directory: %w", err)
	}
	if !info.IsDir() {
		return errors.New("state-dir must be a directory")
	}
	var cfg struct {
		Model    string
		Timeout  string
		MaxTurns int `yaml:"max_turns"`
		Prompt   string
	}
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		return err
	}
	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil || timeout <= 0 || cfg.Model == "" || cfg.MaxTurns < 1 || cfg.Prompt == "" {
		return errors.New("invalid embedded prototype configuration")
	}
	args := []string{"--safe-mode", "--print", "--model", cfg.Model, "--tools", "", "--disable-slash-commands", "--permission-mode", "dontAsk", "--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`, "--setting-sources", "", "--system-prompt", cfg.Prompt, "--json-schema", schema, "--output-format", "json", "--max-turns", strconv.Itoa(cfg.MaxTurns)}
	if resume != "" {
		args = append(args, "--resume", resume)
	}
	if dryRun {
		return json.NewEncoder(os.Stdout).Encode(args)
	}
	input, err := io.ReadAll(io.LimitReader(os.Stdin, 65537))
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(input)) == 0 || len(input) > 65536 {
		return errors.New("stdin must contain 1..65536 bytes of synthetic evidence")
	}
	for _, name := range []string{"config", "work", "tmp"} {
		if err := os.MkdirAll(filepath.Join(state, name), 0700); err != nil {
			return err
		}
	}
	output, err := invoke(state, args, input, timeout)
	if err != nil {
		return err
	}
	parsed, err := parseResult(output, resume)
	if err != nil {
		return err
	}
	return yaml.NewEncoder(os.Stdout).Encode(parsed)
}

func invoke(state string, args []string, input []byte, timeout time.Duration) ([]byte, error) {
	raw, err := os.CreateTemp(state, "response-*.json")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := raw.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "closing raw response:", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = filepath.Join(state, "work")
	cmd.Env = childEnv(state)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), raw, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	fmt.Fprintf(os.Stderr, "Claude safe-mode prototype: timeout=%s; raw response=%s\n", timeout, raw.Name())
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-finished:
			if ctx.Err() != nil {
				return nil, fmt.Errorf("prototype timeout; no accepted verdict: %w", ctx.Err())
			}
			if err != nil {
				return nil, fmt.Errorf("Claude failed; inspect %s: %w", raw.Name(), err)
			}
			if _, err := raw.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			data, err := io.ReadAll(io.LimitReader(raw, 2*1024*1024+1))
			if len(data) > 2*1024*1024 {
				return nil, errors.New("response exceeds 2 MiB")
			}
			return data, err
		case <-ticker.C:
			fmt.Fprintln(os.Stderr, "Claude process still running; this is liveness, not review progress")
		}
	}
}

func childEnv(state string) []string {
	env := make([]string, 0, len(os.Environ())+3)
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if key != "CLAUDE_CONFIG_DIR" && key != "CLAUDE_CODE_TMPDIR" && key != "TMPDIR" {
			env = append(env, item)
		}
	}
	return append(env, "CLAUDE_CONFIG_DIR="+filepath.Join(state, "config"), "CLAUDE_CODE_TMPDIR="+filepath.Join(state, "tmp"), "TMPDIR="+filepath.Join(state, "tmp"))
}

func parseResult(data []byte, resume string) (result, error) {
	var envelope struct {
		Type       string          `json:"type"`
		Subtype    string          `json:"subtype"`
		IsError    bool            `json:"is_error"`
		SessionID  string          `json:"session_id"`
		Structured json.RawMessage `json:"structured_output"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return result{}, fmt.Errorf("invalid result envelope: %w", err)
	}
	if envelope.Type != "result" || envelope.Subtype != "success" || envelope.IsError || strings.TrimSpace(envelope.SessionID) == "" {
		return result{}, errors.New("provider did not return a successful native result")
	}
	if resume != "" && resume != envelope.SessionID {
		return result{}, errors.New("returned session differs from requested resume")
	}
	got := result{SessionID: envelope.SessionID}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Structured))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&got.Response); err != nil {
		return result{}, fmt.Errorf("invalid structured result: %w", err)
	}
	if got.Response.Findings == nil || len(got.Response.Findings) > 10 {
		return result{}, errors.New("findings must be an array with at most ten items")
	}
	switch got.Response.Verdict {
	case "PASS":
		if len(got.Response.Findings) != 0 {
			return result{}, errors.New("PASS must not carry findings")
		}
	case "FAIL", "PROVISIONAL PASS":
		if len(got.Response.Findings) == 0 {
			return result{}, errors.New("non-PASS verdict must explain findings")
		}
	default:
		return result{}, errors.New("unsupported verdict")
	}
	for _, finding := range got.Response.Findings {
		if strings.TrimSpace(finding) == "" {
			return result{}, errors.New("empty finding")
		}
	}
	return got, nil
}
