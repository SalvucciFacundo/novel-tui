package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/model"
)

// version is set at release time via GoReleaser ldflags (-X main.version).
var version = "dev"

func main() {
	// Self-update subcommand: novel-tui upgrade
	if len(os.Args) > 1 && os.Args[1] == "upgrade" {
		if err := runUpgrade(); err != nil {
			fmt.Printf("Error upgrading novel-tui: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Bare `version` alias for `--version`.
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("novel-tui %s\n", effectiveVersion())
		return
	}

	var workspaceDir string
	var showVersion bool
	flag.StringVar(&workspaceDir, "dir", "", "Directory path for novel workspace")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("novel-tui %s\n", effectiveVersion())
		return
	}

	// Also support positional directory argument (e.g. novel-tui ~/Novelas/MiNovela)
	if workspaceDir == "" && flag.NArg() > 0 {
		workspaceDir = flag.Arg(0)
	}

	// A positional argument must be an existing directory. Otherwise a typo
	// (e.g. `novel-tui versoin`) would silently open the editor pointed at a
	// garbage path instead of failing fast.
	if workspaceDir != "" {
		if info, err := os.Stat(repository.ExpandHome(workspaceDir)); err != nil || !info.IsDir() {
			fmt.Printf("Directorio inválido: %s\n\n", workspaceDir)
			flag.Usage()
			os.Exit(2)
		}
	}

	configRepo := repository.NewFileConfigRepository("")
	workspaceMgr := service.NewWorkspaceManager()

	initialView := messages.ViewStateLauncher
	if workspaceDir != "" {
		initialView = messages.ViewStateEditor
	}

	rootModel := model.NewRootModelWithConfig(configRepo, workspaceMgr, initialView, workspaceDir)
	p := tea.NewProgram(rootModel, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running novel-tui: %v\n", err)
		os.Exit(1)
	}
}

// effectiveVersion reports the release version when available. Binaries built
// with GoReleaser carry it via ldflags; binaries from `go install @vX` expose
// it through the module build info; local builds report "dev".
func effectiveVersion() string {
	if version != "" && version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

// runUpgrade reinstalls the latest published novel-tui over the current
// binary using the Go toolchain (the documented install method).
func runUpgrade() error {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("se necesita el toolchain de Go instalado para actualizar (go install)")
	}
	fmt.Printf("Actualizando novel-tui (versión actual: %s)...\n", effectiveVersion())
	cmd := exec.Command(goBin, "install", "github.com/SalvucciFacundo/novel-tui/cmd/novel-tui@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falló go install: %w", err)
	}
	fmt.Println("Listo. Reiniciá novel-tui para usar la nueva versión.")
	return nil
}
