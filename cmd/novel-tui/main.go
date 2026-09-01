package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/model"
)

func main() {
	workspaceDir := flag.String("dir", ".", "Directory path for novel workspace")
	flag.Parse()

	// Initialize repositories
	chapterRepo, err := repository.NewFileChapterRepository(*workspaceDir)
	if err != nil {
		log.Fatalf("Error initializing chapter repository: %v", err)
	}

	characterRepo := repository.NewFileCharacterRepository(*workspaceDir)

	// Ensure workspace setup
	workspaceSvc := service.NewWorkspaceService(*workspaceDir, chapterRepo, characterRepo)
	if err := workspaceSvc.EnsureWorkspace(); err != nil {
		log.Fatalf("Error initializing workspace: %v", err)
	}

	// Create root model and start Bubble Tea program
	rootModel := model.NewRootModel(chapterRepo, characterRepo)
	p := tea.NewProgram(rootModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, an error occurred: %v\n", err)
		os.Exit(1)
	}
}
