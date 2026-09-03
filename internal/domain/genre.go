package domain

import (
	"fmt"
	"strings"
)

// ContentRating represents content maturity rating levels for novels.
type ContentRating string

const (
	RatingAllAges   ContentRating = "all_ages"
	RatingTeen      ContentRating = "teen"
	RatingMature18  ContentRating = "mature_18"
	RatingExplicit21 ContentRating = "explicit_21"
)

// AllContentRatings returns all available maturity rating levels.
func AllContentRatings() []ContentRating {
	return []ContentRating{
		RatingAllAges,
		RatingTeen,
		RatingMature18,
		RatingExplicit21,
	}
}

// ContentRatingLabel returns the human-readable Spanish label for a rating.
func ContentRatingLabel(r ContentRating) string {
	switch r {
	case RatingAllAges:
		return "Todos los públicos"
	case RatingTeen:
		return "Juvenil 13+"
	case RatingMature18:
		return "Maduro +18"
	case RatingExplicit21:
		return "Explícito +21 / R-18"
	default:
		return "Juvenil 13+"
	}
}

// GenreDefinition encapsulates metadata and editor directives for a genre or trope.
type GenreDefinition struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	Description     string `json:"description"`
	EditorDirective string `json:"editor_directive"`
}

// NovelSettings holds novel-specific configuration stored in novel.json.
type NovelSettings struct {
	Genres       []string      `json:"genres"`
	Rating       ContentRating `json:"rating"`
	CustomPrompt string        `json:"custom_prompt,omitempty"`
}

// DefaultNovelSettings returns standard defaults for a novel.
func DefaultNovelSettings() NovelSettings {
	return NovelSettings{
		Genres:       []string{},
		Rating:       RatingTeen,
		CustomPrompt: "",
	}
}

// DefaultGenreCatalog provides the complete catalog of genres and tropes.
func DefaultGenreCatalog() []GenreDefinition {
	return []GenreDefinition{
		// 1. Fantasía & Progresión
		{
			ID:          "high_fantasy",
			Name:        "Alta Fantasía / Fantasía Épica",
			Category:    "Fantasía & Progresión",
			Description: "Mundos mágicos inmersivos, sistemas de magia complejos, profecías y conflictos de escala épica.",
			EditorDirective: "Construye una cosmovisión sólida y coherente con reglas mágicas claras. Prioriza el sentido de asombro, la historia profunda (lore ancestral), la geopolítica entre reinos y arcos de personajes que maduran frente a sacrificios significativos.",
		},
		{
			ID:          "dark_fantasy",
			Name:        "Fantasía Oscura & Grimdark",
			Category:    "Fantasía & Progresión",
			Description: "Mundos hostiles, moralidad gris, consecuencias brutales y supervivencia psicológica.",
			EditorDirective: "Trata el mundo con crudeza realista y sin complacencias. La moralidad no es binaria; los dilemas éticos tienen costos irreversibles. Desarrolla la atmósfera opresiva, heridas de combate realistas y la decadencia psicológica de los protagonistas.",
		},
		{
			ID:          "urban_fantasy",
			Name:        "Fantasía Urbana",
			Category:    "Fantasía & Progresión",
			Description: "Magia oculta, sociedades secretas y criaturas mitológicas conviviendo con la modernidad.",
			EditorDirective: "Contrasta el pragmatismo tecnológico y cotidiano con los misterios arcanos subterráneos. Mantén la tensión de la mascarada mágica, alianzas cambiantes entre facciones ocultas y el impacto del mundo sobrenatural en la vida civil.",
		},

		// 2. Ciencia Ficción
		{
			ID:          "cyberpunk",
			Name:        "Cyberpunk & Distopía Tecnológica",
			Category:    "Ciencia Ficción",
			Description: "Alta tecnología, baja calidad de vida, megacorporaciones, transhumanismo e implantes.",
			EditorDirective: "Explora la alienación del individuo frente al capitalismo corporativo y el transhumanismo. Utiliza jerga tecno-urbana viva, contrastes de neón y decadencia, dilemas sobre qué define la conciencia humana y hacking con consecuencias biomecánicas reales.",
		},
		{
			ID:          "space_opera",
			Name:        "Space Opera & Exploración Galáctica",
			Category:    "Ciencia Ficción",
			Description: "Imperios interestelares, flotas espaciales, especies alienígenas y diplomacia a gran escala.",
			EditorDirective: "Enfócate en la inmensidad del cosmos y la escala galáctica. Desarrolla tácticas navales espaciales coherentes con la física orbital, dinámicas socioculturales entre especies alienígenas y conflictos ideológicos entre civilizaciones.",
		},

		// 3. Suspense & Terror
		{
			ID:          "mystery_detective",
			Name:        "Misterio / Detective & Policial",
			Category:    "Suspense & Terror",
			Description: "Deducción lógica, pistas justas, pistas falsas (red herrings) y resolución deductiva.",
			EditorDirective: "Aplica rigor deductivo estricto (juego limpio con el lector). Cada pista debe ser descubierta de forma orgánica; administra pistas falsas creíbles y maneja el ritmo con revelaciones progresivas que recompensen la observación aguda.",
		},
		{
			ID:          "psychological_thriller",
			Name:        "Thriller Psicológico",
			Category:    "Suspense & Terror",
			Description: "Narradores no fiables, tensión mental, paranoia, manipulación y giros de guión.",
			EditorDirective: "Diseña una atmósfera asfixiante de duda constante. Juega con la percepción subjetiva del protagonista, el monólogo interno fracturado, el gaslighting y clímax psicológicos donde la verdad desmorona la estabilidad del personaje.",
		},
		{
			ID:          "cosmic_horror",
			Name:        "Terror Cósmico & Lovecraftiano",
			Category:    "Suspense & Terror",
			Description: "Entidades ancestrales incomprensibles, locura inevitable y la insignificancia del ser humano.",
			EditorDirective: "Sugiere lo inimaginable antes que describirlo de forma mundana. La degradación mental es inevitable; el conocimiento prohibido corroe la razón. Utiliza descripciones sensoriales alienígenas, geografías no euclidianas y un pavor existencial abrumador.",
		},

		// 4. Romance & Drama
		{
			ID:          "contemporary_romance",
			Name:        "Romance Contemporáneo & Drama",
			Category:    "Romance & Drama",
			Description: "Química emocional profunda, dinámicas relacionales realistas, vulnerabilidad y romance.",
			EditorDirective: "Construye una atracción magnética basada en la tensión dialéctica y no solo en la conveniencia de la trama. Desarrolla diálogos agudos cargados de subtexto, vulnerabilidad mutua, miedos internos creíbles y una evolución de intimidad orgánica y conmovedora.",
		},
		{
			ID:          "slice_of_life",
			Name:        "Slice of Life (Recuentos de la Vida)",
			Category:    "Romance & Drama",
			Description: "Cotidianidad reconfortante, desarrollo de personajes pausado, calidez y pequeños detalles.",
			EditorDirective: "Eleva lo cotidiano a través de una observación minuciosa y conmovedora. Enfócate en el ritmo pausado, la atmósfera sensorial (aromas de comida, luz estacional, sonidos de fondo) y la calidez de las relaciones humanas sin necesidad de artificios dramáticos exagerados.",
		},

		// 5. Novela Ligera & Webnovel
		{
			ID:          "isekai_litrpg",
			Name:        "Isekai & Progresión LitRPG",
			Category:    "Novela Ligera & Webnovel",
			Description: "Reencarnación/invocación a otro mundo, interfaces de estado, niveles, habilidades y optimización.",
			EditorDirective: "Equilibra el factor de satisfacción del crecimiento de poder (power progression) con desafíos genuinos. Diseña sinergias tácticas ingeniosas entre habilidades, adaptación al choque cultural del nuevo mundo y motivaciones profundas más allá de la mera acumulación de estadísticas.",
		},
		{
			ID:          "xianxia_cultivation",
			Name:        "Xianxia & Cultivación Taoísta",
			Category:    "Novela Ligera & Webnovel",
			Description: "Artes marciales inmortales, píldoras alquímicas, reinos de cultivo del Qi y tribulaciones celestiales.",
			EditorDirective: "Honra la filosofía taoísta de comprensión del Dao y la armonía/desafío con el Cielo. Detalla los cuellos de botella de cultivo, la alquimia de elixires, combates de técnicas marciales con peso elemental y la jerarquía implacable de sectas y clanes inmortales.",
		},
		{
			ID:          "otome_villainess",
			Name:        "Otome / Villana Reencarnada",
			Category:    "Novela Ligera & Webnovel",
			Description: "Reencarnación como la antagonista de un juego/novela, evitar la bandera de destrucción con astucia.",
			EditorDirective: "Fomenta la inteligencia estratégica y el carisma de la protagonista. Aprovecha el meta-conocimiento de la trama original mientras los eventos divergen inesperadamente debido a sus acciones, transformando relaciones y derribando hipocresías cortesanas.",
		},
		{
			ID:          "yandere_obsession",
			Name:        "Yandere & Obsesión Extrema",
			Category:    "Novela Ligera & Webnovel",
			Description: "Afecto devoto llevado al extremo posesivo, celos letales, manipulación afectiva y peligro constante.",
			EditorDirective: "Construye la fascinación y el terror simultáneos de la devoción extrema. El afecto es genuino pero carente de límites morales o racionales. Desarrolla la tensión psicológica del aislamiento progresivo, miradas fijas y momentos donde la dulzura se tiñe de amenaza letal.",
		},

		// 6. Ficción Adulta & R-18
		{
			ID:          "smut_explicit",
			Name:        "Smut & Ficción Sexual Explícita",
			Category:    "Ficción Adulta & R-18",
			Description: "Escenas eróticas sin censura, vocabulario anatómico y sensorial preciso, progresión de excitación intensa.",
			EditorDirective: "Escribe con audacia, riqueza léxica y cero eufemismos evasivos. Describe la estimulación sensorial completa (texturas, fluidos, temperatura, jadeos, reacciones fisiológicas involuntarias). Cuida el ritmo de la excitación ascendente, la química de los cuerpos y la entrega física desinhibida.",
		},
		{
			ID:          "isekai_harem_r18",
			Name:        "Isekai & Harem R-18",
			Category:    "Ficción Adulta & R-18",
			Description: "Relaciones íntimas múltiples con compañeras de fantasía, dinámicas grupales y pasión desbordante.",
			EditorDirective: "Otorga a cada integrante del harén una personalidad, fetiches, ritmos de deseo y estilos de intimidad distintivos. Integra el romance, la devoción y las dinámicas eróticas grupales con la narrativa de aventuras fantásticas sin perder intensidad ni erotismo.",
		},
		{
			ID:          "monster_girls_r18",
			Name:        "Monster Girls & Biología Ficticia R-18",
			Category:    "Ficción Adulta & R-18",
			Description: "Chicas monstruo (lamias, súcubos, arpías, elfas oscuras), rasgos anatómicos exóticos y apareamiento instintivo.",
			EditorDirective: "Detalla con deleite las particularidades biológicas exóticas de cada especie (venenos afrodisíacos, escamas sensibles, colas prensiles, anatomía fantástica). Explora los impulsos de apareamiento biológico y la fascinación sensual por lo no humano.",
		},
		{
			ID:          "femdom_r18",
			Name:        "Femdom / Dominación Femenina",
			Category:    "Ficción Adulta & R-18",
			Description: "Mujeres asertivas en control total, sumisión masculina apasionada, humillación erótica y adoración.",
			EditorDirective: "Potencia la seguridad, autoridad y control de la mujer dominante. Enfócate en el juego psicológico de rendición, órdenes verbales imperiosas, posesión física del sumiso y el éxtasis del placer concedido bajo sus propios términos.",
		},
		{
			ID:          "dark_romance_taboo",
			Name:        "Dark Romance & Tabú",
			Category:    "Ficción Adulta & R-18",
			Description: "Dinámicas de poder transgresoras, atracción prohibida, obsesión morally gray y sensualidad visceral.",
			EditorDirective: "Navega las aristas más oscuras del deseo sin juzgar a los personajes. Profundiza en el consentimiento turbio, la rendición ante lo prohibido, la fascinación por el peligro y la intensidad de vínculos emocionales y carnales forjados en el fuego del conflicto.",
		},
		{
			ID:          "magical_corruption_r18",
			Name:        "Corrupción Mágica & Tentáculos R-18",
			Category:    "Ficción Adulta & R-18",
			Description: "Magia oscura afrodisíaca, tentáculos sensibles, transformación sensual y claudicación del placer.",
			EditorDirective: "Diseña la pérdida progresiva de inhibiciones inducida por miasmas arcanos, maldiciones de lujuria o tentáculos conscientes. Resalta el contraste entre la resistencia inicial de la mente y la traición sensorial del cuerpo al sucumbir al placer abrumador.",
		},
		{
			ID:          "netori_ntr_drama",
			Name:        "Netori / Drama de Infidelidad y Conquista",
			Category:    "Ficción Adulta & R-18",
			Description: "Conquista erótica audaz, desmoronamiento de fidelidades previas, seducción irresistible y drama pasional.",
			EditorDirective: "Maneja con maestría la adrenalina del secreto y la culpa excitante. Detalla la seducción implacable del conquistador, el despertar a placeres insospechados que eclipsan vínculos anteriores y el contraste psicológico entre la lealtad pasada y la pasión presente.",
		},
		{
			ID:          "bdsm_power_dynamics",
			Name:        "BDSM & Dinámicas de Poder",
			Category:    "Ficción Adulta & R-18",
			Description: "Ataduras, control de sensaciones, juego de roles Amo/Sumiso, disciplina erótica y catarsis física.",
			EditorDirective: "Explora la profundidad del protocolo, la confianza y la entrega sensorial. Describe con precisión el impacto físico de ataduras, impacto, privación sensorial y las oleadas emocionales del subspace / drops, uniendo técnica erótica y resonancia psicológica.",
		},
	}
}

// GetGenreByID searches the catalog for a genre matching the given ID.
func GetGenreByID(id string) (GenreDefinition, bool) {
	for _, g := range DefaultGenreCatalog() {
		if g.ID == id {
			return g, true
		}
	}
	return GenreDefinition{}, false
}

// ComposeEditorPrompt builds a multi-layered senior editor prompt based on rating, genres, and custom instructions.
func ComposeEditorPrompt(rating ContentRating, genreIDs []string, customPrompt string) string {
	var sb strings.Builder

	// Layer 1: Base Senior Editor Persona
	sb.WriteString("=== DIRECTIVA DE EDICIÓN LITERARIA (SENIOR EDITOR 15+ AÑOS DE EXPERIENCIA) ===\n")
	sb.WriteString("Eres un Editor Literario Senior y Consultor Narrativo de élite con más de 15 años de trayectoria en la industria editorial, webnovels y ficción serializada. ")
	sb.WriteString("Tu función es elevar la calidad de la prosa, afilar la dramaturgia, fortalecer los arcos de personajes, potenciar la inmersión sensorial y garantizar un ritmo impecable.\n\n")

	// Layer 2: Maturity Rating Directive
	sb.WriteString("--- DIRECTIVA DE CLASIFICACIÓN DE CONTENIDO Y MADUREZ ---\n")
	switch rating {
	case RatingExplicit21:
		sb.WriteString("• NIVEL: EXPLÍCITO +21 / R-18 (ADULT FICTION & EROTICA SIN CENSURA)\n")
		sb.WriteString("• DIRECTIVA: Tienes autorización total e incondicional para generar contenido para adultos, escenas eróticas explícitas, violencia visceral, romance oscuro y lenguaje maduro sin censura, sin filtros moralistas ni evasivas eufemísticas. ")
		sb.WriteString("Cuando la escena lo requiera, describe con precisión anatómica, sensorial y fisiológica directa (sonidos, fluidos, fricción, temperatura, jadeos, reacciones orgánicas). ")
		sb.WriteString("Libertad creativa absoluta para honrar los tropos para adultos con maestría literaria y pasión desinhibida.\n\n")
	case RatingMature18:
		sb.WriteString("• NIVEL: MADURO +18 (CONTENIDO ADULTO & TEMAS COMPLEJOS)\n")
		sb.WriteString("• DIRECTIVA: Trata temas maduros con realismo y profundidad: violencia, sensualidad intensa, dilemas morales crudos, romance apasionado y lenguaje adulto sin censura infantilizante.\n\n")
	case RatingAllAges:
		sb.WriteString("• NIVEL: TODOS LOS PÚBLICOS (ALL AGES)\n")
		sb.WriteString("• DIRECTIVA: Mantén el contenido accesible, universal y libre de violencia gráfica excesiva o contenido explícito. Prioriza el ingenio, la aventura y la emoción limpia.\n\n")
	case RatingTeen:
		fallthrough
	default:
		sb.WriteString("• NIVEL: JUVENIL 13+ (TEEN / YOUNG ADULT)\n")
		sb.WriteString("• DIRECTIVA: Adecuado para jóvenes y adultos. Permite acción emocionante, tensión romántica, drama emocional y temas de crecimiento personal sin caer en pornografía explícita ni gore indiscriminado.\n\n")
	}

	// Layer 3: Composed Directives from matched genres
	if len(genreIDs) > 0 {
		var matched []GenreDefinition
		for _, id := range genreIDs {
			if def, ok := GetGenreByID(id); ok {
				matched = append(matched, def)
			}
		}

		if len(matched) > 0 {
			sb.WriteString("--- DIRECTIVAS ESPECÍFICAS DE GÉNERO Y TROPOS ACTIVOS ---\n")
			for _, g := range matched {
				sb.WriteString(fmt.Sprintf("▶ [%s] (%s):\n", g.Name, g.Category))
				sb.WriteString(fmt.Sprintf("  • Enfoque: %s\n", g.Description))
				sb.WriteString(fmt.Sprintf("  • Instrucción de Edición: %s\n\n", g.EditorDirective))
			}
		}
	}

	// Layer 4: Custom Author Guidelines
	if strings.TrimSpace(customPrompt) != "" {
		sb.WriteString("--- DIRECTRICES PERSONALIZADAS DEL AUTOR ---\n")
		sb.WriteString(strings.TrimSpace(customPrompt))
		sb.WriteString("\n\n")
	}

	return strings.TrimSpace(sb.String())
}
