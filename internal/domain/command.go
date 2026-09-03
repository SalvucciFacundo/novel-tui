package domain

// CommandItem represents an executable action or shortcut displayed in the Command Palette.
type CommandItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Shortcut    string `json:"shortcut"`
	Description string `json:"description"`
}

// DefaultCommands returns the standard registry of command palette actions and shortcuts.
func DefaultCommands() []CommandItem {
	return []CommandItem{
		{
			ID:          "global_search",
			Title:       "Búsqueda y Reemplazo Global",
			Category:    "Editor",
			Shortcut:    "Ctrl+F",
			Description: "Buscar o reemplazar texto en todos los capítulos",
		},
		{
			ID:          "save_chapter",
			Title:       "Guardar Capítulo",
			Category:    "Editor",
			Shortcut:    "Ctrl+S",
			Description: "Guardar el capítulo actual",
		},
		{
			ID:          "new_chapter",
			Title:       "Nuevo Capítulo",
			Category:    "Editor",
			Shortcut:    "Ctrl+N",
			Description: "Crear un nuevo capítulo en la novela",
		},
		{
			ID:          "toggle_ai",
			Title:       "Alternar Asistente IA",
			Category:    "Asistente IA",
			Shortcut:    "Ctrl+A",
			Description: "Abrir o cerrar el panel lateral del asistente IA",
		},
		{
			ID:          "go_launcher",
			Title:       "Volver al Inicio",
			Category:    "Navegación",
			Shortcut:    "Ctrl+H",
			Description: "Regresar a la pantalla de inicio y selección de novelas",
		},
		{
			ID:          "tab_chapters",
			Title:       "Ir a Capítulos",
			Category:    "Navegación",
			Shortcut:    "1",
			Description: "Mostrar la pestaña de capítulos en la barra lateral",
		},
		{
			ID:          "tab_characters",
			Title:       "Ir a Personajes",
			Category:    "Navegación",
			Shortcut:    "2",
			Description: "Mostrar la pestaña de personajes",
		},
		{
			ID:          "tab_notes",
			Title:       "Ir a Notas",
			Category:    "Navegación",
			Shortcut:    "3",
			Description: "Mostrar la pestaña de notas del proyecto",
		},
		{
			ID:          "tab_brain",
			Title:       "Ir a Memoria Brain",
			Category:    "Memoria Brain",
			Shortcut:    "4",
			Description: "Ver la memoria semántica y entidades de la novela",
		},
		{
			ID:          "toggle_timeline",
			Title:       "Alternar Línea Temporal",
			Category:    "Memoria Brain",
			Shortcut:    "t",
			Description: "Alternar entre vista de entidades y línea temporal en Brain",
		},
		{
			ID:          "llm_config",
			Title:       "Configuración de IA",
			Category:    "Configuración",
			Shortcut:    "c",
			Description: "Configurar proveedor, modelos y claves de API de IA",
		},
	}
}
