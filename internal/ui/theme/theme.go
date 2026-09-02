package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Palette defines the colors used throughout the application.
type Palette struct {
	Background    lipgloss.Color
	Foreground    lipgloss.Color
	BorderFocused lipgloss.Color
	BorderBlurred lipgloss.Color
	Accent        lipgloss.Color
	Secondary     lipgloss.Color
	Muted         lipgloss.Color
	Highlight     lipgloss.Color
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Error         lipgloss.Color
	CardBg        lipgloss.Color
}

// CatppuccinMocha palette definition
var CatppuccinMocha = Palette{
	Background:    lipgloss.Color("#1e1e2e"),
	Foreground:    lipgloss.Color("#cdd6f4"),
	BorderFocused: lipgloss.Color("#cba6f7"), // Mauve
	BorderBlurred: lipgloss.Color("#45475a"), // Surface1
	Accent:        lipgloss.Color("#89b4fa"), // Blue
	Secondary:     lipgloss.Color("#b4befe"), // Lavender
	Muted:         lipgloss.Color("#6c7086"), // Overlay0
	Highlight:     lipgloss.Color("#f5c2e7"), // Pink
	Success:       lipgloss.Color("#a6e3a1"), // Green
	Warning:       lipgloss.Color("#fab387"), // Peach
	Error:         lipgloss.Color("#f38ba8"), // Red
	CardBg:        lipgloss.Color("#313244"), // Surface0
}

// CurrentTheme is the active palette
var CurrentTheme = CatppuccinMocha

// Styles holds all pre-configured Lip Gloss styles.
type Styles struct {
	AppContainer lipgloss.Style

	// Panel borders
	FocusedPanel lipgloss.Style
	BlurredPanel lipgloss.Style

	// Sidebar styles
	SidebarHeader  lipgloss.Style
	TabActive      lipgloss.Style
	TabInactive    lipgloss.Style
	ListItem       lipgloss.Style
	ListItemActive lipgloss.Style
	ListSubtitle   lipgloss.Style

	// Character card styles
	CardContainer   lipgloss.Style
	CardName        lipgloss.Style
	CardRole        lipgloss.Style
	CardDescription lipgloss.Style
	CardNotes       lipgloss.Style

	// Editor styles
	EditorContainer lipgloss.Style

	// Status Bar styles
	StatusBarContainer lipgloss.Style
	StatusTitle        lipgloss.Style
	StatusSaved        lipgloss.Style
	StatusDirty        lipgloss.Style
	StatusMetrics      lipgloss.Style
	StatusHint         lipgloss.Style
	StatusCommandBadge lipgloss.Style

	// Navbar styles
	NavbarContainer lipgloss.Style
	NavbarPill      lipgloss.Style
	NavbarHomePill  lipgloss.Style
	NavbarBreadcrumb lipgloss.Style
	NavbarActionPill lipgloss.Style

	// Warning view
	WarningView lipgloss.Style
}

// NewStyles constructs the UI styling rules based on the given palette.
func NewStyles(p Palette) Styles {
	return Styles{
		AppContainer: lipgloss.NewStyle().
			Background(p.Background),

		FocusedPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.BorderFocused),

		BlurredPanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.BorderBlurred),

		SidebarHeader: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1),

		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Highlight).
			Background(p.CardBg).
			Padding(0, 1),

		TabInactive: lipgloss.NewStyle().
			Foreground(p.Muted).
			Padding(0, 1),

		ListItem: lipgloss.NewStyle().
			Foreground(p.Foreground).
			Padding(0, 1),

		ListItemActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Background).
			Background(p.BorderFocused).
			Padding(0, 1),

		ListSubtitle: lipgloss.NewStyle().
			Foreground(p.Muted).
			PaddingLeft(2),

		CardContainer: lipgloss.NewStyle().
			Background(p.CardBg).
			Padding(1, 1).
			MarginTop(1),

		CardName: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Secondary),

		CardRole: lipgloss.NewStyle().
			Foreground(p.Accent).
			Italic(true),

		CardDescription: lipgloss.NewStyle().
			Foreground(p.Foreground).
			MarginTop(1),

		CardNotes: lipgloss.NewStyle().
			Foreground(p.Muted).
			MarginTop(1),

		EditorContainer: lipgloss.NewStyle().
			Padding(0, 1),

		StatusBarContainer: lipgloss.NewStyle().
			Background(p.CardBg).
			Foreground(p.Foreground).
			Height(1),

		StatusTitle: lipgloss.NewStyle().
			Bold(true).
			Background(p.Accent).
			Foreground(p.Background).
			Padding(0, 1),

		StatusSaved: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Success).
			Padding(0, 1),

		StatusDirty: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Warning).
			Padding(0, 1),

		StatusMetrics: lipgloss.NewStyle().
			Foreground(p.Foreground).
			Padding(0, 1),

		StatusHint: lipgloss.NewStyle().
			Foreground(p.Muted).
			Padding(0, 1),

		StatusCommandBadge: lipgloss.NewStyle().
			Foreground(p.Foreground).
			Background(p.Background).
			Padding(0, 1),

		NavbarContainer: lipgloss.NewStyle().
			Background(p.CardBg).
			Foreground(p.Foreground).
			Height(1),

		NavbarPill: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Background(p.CardBg).
			Padding(0, 1),

		NavbarHomePill: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Accent).
			Background(p.Background).
			Padding(0, 1),

		NavbarBreadcrumb: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Highlight),

		NavbarActionPill: lipgloss.NewStyle().
			Foreground(p.Foreground).
			Background(p.Background).
			Padding(0, 1),

		WarningView: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Warning).
			Align(lipgloss.Center, lipgloss.Center),
	}
}

// DefaultStyles provides access to global default styles.
var DefaultStyles = NewStyles(CurrentTheme)
