package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ColabService handles automated provisioning and lifecycle of the Google Colab GPU LLM server.
type ColabService interface {
	CheckColabCLIInstalled() bool
	StartColabServer(ctx context.Context) (baseURL string, err error)
	StopColabServer(ctx context.Context) error
}

// CommandRunner defines the function signature for executing external system commands with context.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// LookPathFunc defines the function signature for locating binaries in system PATH.
type LookPathFunc func(file string) (string, error)

type defaultColabService struct {
	lookPath   LookPathFunc
	runner     CommandRunner
	scriptPath string
}

var cloudflareRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com(?:/v1)?`)

// NewColabService creates a default ColabService using system PATH and standard exec.
func NewColabService() ColabService {
	return &defaultColabService{
		lookPath: exec.LookPath,
		runner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			return cmd.CombinedOutput()
		},
	}
}

// NewColabServiceWithRunner creates a ColabService with custom command execution hooks for testing.
func NewColabServiceWithRunner(lookPath LookPathFunc, runner CommandRunner) ColabService {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if runner == nil {
		runner = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			return cmd.CombinedOutput()
		}
	}
	return &defaultColabService{
		lookPath: lookPath,
		runner:   runner,
	}
}

// CheckColabCLIInstalled returns true if the 'colab' executable is accessible in PATH.
func (s *defaultColabService) CheckColabCLIInstalled() bool {
	if s.lookPath == nil {
		return false
	}
	_, err := s.lookPath("colab")
	return err == nil
}

// StartColabServer provisions or runs the Google Colab LLM server and returns the Cloudflare API baseURL.
func (s *defaultColabService) StartColabServer(ctx context.Context) (string, error) {
	// Determine command to run
	name := "python3"
	args := []string{"scripts/colab_server.py"}

	// If scripts/start_colab_llm.sh exists in cwd or relative path, check executable or use python
	if s.scriptPath != "" {
		args = []string{s.scriptPath}
	} else if _, err := os.Stat("scripts/start_colab_llm.sh"); err == nil {
		name = "bash"
		args = []string{"scripts/start_colab_llm.sh"}
	} else if _, err := os.Stat("scripts/colab_server.py"); err == nil {
		name = "python3"
		args = []string{"scripts/colab_server.py"}
	} else {
		// Fallback to colab CLI directly if script files are not found in current directory
		name = "colab"
		args = []string{"run", "--session", "novel-llm"}
	}

	outBytes, err := s.runner(ctx, name, args...)
	outStr := string(outBytes)

	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if len(outBytes) > 0 {
			return "", fmt.Errorf("colab execution failed: %s (%w)", strings.TrimSpace(outStr), err)
		}
		return "", fmt.Errorf("colab execution failed: %w", err)
	}

	match := cloudflareRegex.FindString(outStr)
	if match == "" {
		return "", fmt.Errorf("no Cloudflare tunnel URL found in Colab output:\n%s", outStr)
	}

	match = strings.TrimSpace(match)
	if !strings.HasSuffix(match, "/v1") {
		match = strings.TrimRight(match, "/") + "/v1"
	}

	return match, nil
}

// StopColabServer terminates the remote Google Colab session.
func (s *defaultColabService) StopColabServer(ctx context.Context) error {
	name := "colab"
	args := []string{"stop", "--session", "novel-llm"}

	_, err := s.runner(ctx, name, args...)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("failed to stop colab session: %w", err)
	}
	return nil
}
