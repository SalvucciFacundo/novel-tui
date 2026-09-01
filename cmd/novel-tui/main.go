package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/model"
)

func main() {
	var workspaceDir string
	flag.StringVar(&workspaceDir, "dir", "", "Directory path for novel workspace")
	flag.Parse()

	// Also support positional directory argument (e.g. novel-tui ~/Novelas/MiNovela)
	if workspaceDir == "" && flag.NArg() > 0 {
		workspaceDir = flag.Arg(0)
	}

	configRepo := repository.NewFileConfigRepository("")
	workspaceMgr := service.NewWorkspaceManager()

	initialView := messages.ViewStateLauncher
	if workspaceDir != "" {
		initialView = messages.ViewStateEditor
	}

	rootModel := model.NewRootModelWithConfig(configRepo, workspaceMgr, initialView, workspaceDir)
	p := tea.NewProgram(rootModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running novel-tui: %v\n", err)
		os.Exit(1)
	}
}
