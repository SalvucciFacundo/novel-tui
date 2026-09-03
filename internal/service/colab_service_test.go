package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/service"
)

func TestColabService_CheckColabCLIInstalled(t *testing.T) {
	t.Run("CLI installed", func(t *testing.T) {
		svc := service.NewColabServiceWithRunner(
			func(file string) (string, error) {
				if file == "colab" {
					return "/usr/local/bin/colab", nil
				}
				return "", errors.New("not found")
			},
			nil,
		)
		if !svc.CheckColabCLIInstalled() {
			t.Errorf("expected CheckColabCLIInstalled to return true")
		}
	})

	t.Run("CLI not installed", func(t *testing.T) {
		svc := service.NewColabServiceWithRunner(
			func(file string) (string, error) {
				return "", errors.New("executable file not found in $PATH")
			},
			nil,
		)
		if svc.CheckColabCLIInstalled() {
			t.Errorf("expected CheckColabCLIInstalled to return false")
		}
	})
}

func TestColabService_StartColabServer_Success(t *testing.T) {
	mockOutput := `
[Colab] Initializing T4 GPU session 'novel-llm'...
[Colab] Downloading KoboldCpp...
[Colab] Downloading Stheno 8B GGUF...
[KoboldCpp] Initialized with 33 GPU layers on CUDA.
[Cloudflare] Tunnel established: https://quiet-river-1234.trycloudflare.com
[KoboldCpp] OpenAI-compatible endpoint ready at: https://quiet-river-1234.trycloudflare.com/v1
`
	svc := service.NewColabServiceWithRunner(
		func(file string) (string, error) {
			return "/usr/bin/colab", nil
		},
		func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte(mockOutput), nil
		},
	)

	baseURL, err := svc.StartColabServer(context.Background())
	if err != nil {
		t.Fatalf("unexpected error starting colab server: %v", err)
	}

	expectedURL := "https://quiet-river-1234.trycloudflare.com/v1"
	if baseURL != expectedURL {
		t.Errorf("expected baseURL %q, got %q", expectedURL, baseURL)
	}
}

func TestColabService_StartColabServer_URLWithoutV1(t *testing.T) {
	mockOutput := `
Tunnel launched at https://fast-wind-9876.trycloudflare.com
Server is listening!
`
	svc := service.NewColabServiceWithRunner(
		func(file string) (string, error) {
			return "/usr/bin/colab", nil
		},
		func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte(mockOutput), nil
		},
	)

	baseURL, err := svc.StartColabServer(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://fast-wind-9876.trycloudflare.com/v1"
	if baseURL != expectedURL {
		t.Errorf("expected baseURL %q, got %q", expectedURL, baseURL)
	}
}

func TestColabService_StartColabServer_CommandError(t *testing.T) {
	svc := service.NewColabServiceWithRunner(
		func(file string) (string, error) {
			return "/usr/bin/colab", nil
		},
		func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte("Error: GPU quota exceeded or authentication failed"), errors.New("exit status 1")
		},
	)

	baseURL, err := svc.StartColabServer(context.Background())
	if err == nil {
		t.Fatalf("expected error when command fails, got baseURL %q", baseURL)
	}
	if !strings.Contains(err.Error(), "GPU quota exceeded") && !strings.Contains(err.Error(), "exit status 1") {
		t.Errorf("expected error message to contain details, got: %v", err)
	}
}

func TestColabService_StartColabServer_NoURLFound(t *testing.T) {
	svc := service.NewColabServiceWithRunner(
		func(file string) (string, error) {
			return "/usr/bin/colab", nil
		},
		func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte("Session started but no tunnel output generated"), nil
		},
	)

	_, err := svc.StartColabServer(context.Background())
	if err == nil {
		t.Fatalf("expected error when no Cloudflare URL found")
	}
	if !strings.Contains(err.Error(), "cloudflare") && !strings.Contains(err.Error(), "URL") {
		t.Errorf("expected error to mention cloudflare or URL, got: %v", err)
	}
}

func TestColabService_StartColabServer_ContextTimeout(t *testing.T) {
	svc := service.NewColabServiceWithRunner(
		func(file string) (string, error) {
			return "/usr/bin/colab", nil
		},
		func(ctx context.Context, name string, args ...string) ([]byte, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return []byte("https://late.trycloudflare.com/v1"), nil
			}
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := svc.StartColabServer(ctx)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("expected deadline exceeded, got: %v", err)
	}
}

func TestColabService_StopColabServer(t *testing.T) {
	stopped := false
	svc := service.NewColabServiceWithRunner(
		func(file string) (string, error) {
			return "/usr/bin/colab", nil
		},
		func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "stop" {
				stopped = true
				return []byte("Session 'novel-llm' stopped."), nil
			}
			return nil, nil
		},
	)

	err := svc.StopColabServer(context.Background())
	if err != nil {
		t.Fatalf("unexpected error stopping server: %v", err)
	}
	if !stopped {
		t.Errorf("expected stop command to be executed")
	}
}
