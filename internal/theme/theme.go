package theme

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name            string
	Background      lipgloss.Color
	Foreground      lipgloss.Color
	PanelBorder     lipgloss.Color
	ActiveBorder    lipgloss.Color
	HeaderBg        lipgloss.Color
	HeaderFg        lipgloss.Color
	TabActive       lipgloss.Color
	TabInactive     lipgloss.Color
	MethodGet       lipgloss.Color
	MethodPost      lipgloss.Color
	MethodPut       lipgloss.Color
	MethodPatch     lipgloss.Color
	MethodDelete    lipgloss.Color
	Status2xx       lipgloss.Color
	Status3xx       lipgloss.Color
	Status4xx       lipgloss.Color
	Status5xx       lipgloss.Color
	ChromaStyleName string
}

var (
	DarkTheme = Theme{
		Name:            "dark",
		Background:      lipgloss.Color("#1e1e2e"),
		Foreground:      lipgloss.Color("#cdd6f4"),
		PanelBorder:     lipgloss.Color("#45475a"),
		ActiveBorder:    lipgloss.Color("#cba6f7"), // Mauve/Purple
		HeaderBg:        lipgloss.Color("#313244"),
		HeaderFg:        lipgloss.Color("#f5e0dc"),
		TabActive:       lipgloss.Color("#cba6f7"),
		TabInactive:     lipgloss.Color("#6c7086"),
		MethodGet:       lipgloss.Color("#a6e3a1"), // Green
		MethodPost:      lipgloss.Color("#f9e2af"), // Yellow
		MethodPut:       lipgloss.Color("#89b4fa"), // Blue
		MethodPatch:     lipgloss.Color("#94e2d5"), // Cyan
		MethodDelete:    lipgloss.Color("#f38ba8"), // Red
		Status2xx:       lipgloss.Color("#a6e3a1"),
		Status3xx:       lipgloss.Color("#89b4fa"),
		Status4xx:       lipgloss.Color("#f9e2af"),
		Status5xx:       lipgloss.Color("#f38ba8"),
		ChromaStyleName: "dracula",
	}

	LightTheme = Theme{
		Name:            "light",
		Background:      lipgloss.Color("#eff1f5"),
		Foreground:      lipgloss.Color("#4c4f69"),
		PanelBorder:     lipgloss.Color("#bcc0cc"),
		ActiveBorder:    lipgloss.Color("#8839ef"),
		HeaderBg:        lipgloss.Color("#e6e9ef"),
		HeaderFg:        lipgloss.Color("#4c4f69"),
		TabActive:       lipgloss.Color("#8839ef"),
		TabInactive:     lipgloss.Color("#9ca0b0"),
		MethodGet:       lipgloss.Color("#40a02b"),
		MethodPost:      lipgloss.Color("#df8e1d"),
		MethodPut:       lipgloss.Color("#1e66f5"),
		MethodPatch:     lipgloss.Color("#179299"),
		MethodDelete:    lipgloss.Color("#d20f39"),
		Status2xx:       lipgloss.Color("#40a02b"),
		Status3xx:       lipgloss.Color("#1e66f5"),
		Status4xx:       lipgloss.Color("#df8e1d"),
		Status5xx:       lipgloss.Color("#d20f39"),
		ChromaStyleName: "github",
	}

	HighContrastTheme = Theme{
		Name:            "highcontrast",
		Background:      lipgloss.Color("#000000"),
		Foreground:      lipgloss.Color("#ffffff"),
		PanelBorder:     lipgloss.Color("#ffffff"),
		ActiveBorder:    lipgloss.Color("#ffff00"), // Bright Yellow
		HeaderBg:        lipgloss.Color("#111111"),
		HeaderFg:        lipgloss.Color("#ffffff"),
		TabActive:       lipgloss.Color("#ffff00"),
		TabInactive:     lipgloss.Color("#888888"),
		MethodGet:       lipgloss.Color("#00ff00"),
		MethodPost:      lipgloss.Color("#ffff00"),
		MethodPut:       lipgloss.Color("#00ffff"),
		MethodPatch:     lipgloss.Color("#00ffff"),
		MethodDelete:    lipgloss.Color("#ff0000"),
		Status2xx:       lipgloss.Color("#00ff00"),
		Status3xx:       lipgloss.Color("#00ffff"),
		Status4xx:       lipgloss.Color("#ffff00"),
		Status5xx:       lipgloss.Color("#ff0000"),
		ChromaStyleName: "monokai",
	}
	CyberpunkTheme = Theme{
		Name:            "cyberpunk",
		Background:      lipgloss.Color("#0f0e17"),
		Foreground:      lipgloss.Color("#fffffe"),
		PanelBorder:     lipgloss.Color("#ff8906"),
		ActiveBorder:    lipgloss.Color("#ff007f"), // Neon Pink
		HeaderBg:        lipgloss.Color("#2e2f3e"),
		HeaderFg:        lipgloss.Color("#00f0ff"), // Neon Cyan
		TabActive:       lipgloss.Color("#ff007f"),
		TabInactive:     lipgloss.Color("#a7a9be"),
		MethodGet:       lipgloss.Color("#00f0ff"),
		MethodPost:      lipgloss.Color("#ff8906"),
		MethodPut:       lipgloss.Color("#f25f4c"),
		MethodPatch:     lipgloss.Color("#e53170"),
		MethodDelete:    lipgloss.Color("#ff007f"),
		Status2xx:       lipgloss.Color("#00f0ff"),
		Status3xx:       lipgloss.Color("#ff8906"),
		Status4xx:       lipgloss.Color("#f25f4c"),
		Status5xx:       lipgloss.Color("#ff007f"),
		ChromaStyleName: "monokai",
	}

	NordTheme = Theme{
		Name:            "nord",
		Background:      lipgloss.Color("#2e3440"),
		Foreground:      lipgloss.Color("#eceff4"),
		PanelBorder:     lipgloss.Color("#4c566a"),
		ActiveBorder:    lipgloss.Color("#88c0d0"), // Frost Cyan
		HeaderBg:        lipgloss.Color("#3b4252"),
		HeaderFg:        lipgloss.Color("#e5e9f0"),
		TabActive:       lipgloss.Color("#88c0d0"),
		TabInactive:     lipgloss.Color("#d8dee9"),
		MethodGet:       lipgloss.Color("#a3be8c"), // Green
		MethodPost:      lipgloss.Color("#ebcb8b"), // Yellow
		MethodPut:       lipgloss.Color("#81a1c1"), // Blue
		MethodPatch:     lipgloss.Color("#8fbcbb"), // Teal
		MethodDelete:    lipgloss.Color("#bf616a"), // Red
		Status2xx:       lipgloss.Color("#a3be8c"),
		Status3xx:       lipgloss.Color("#81a1c1"),
		Status4xx:       lipgloss.Color("#ebcb8b"),
		Status5xx:       lipgloss.Color("#bf616a"),
		ChromaStyleName: "nord",
	}
)

var AvailableThemes = []string{"dark", "light", "cyberpunk", "nord", "highcontrast"}

func GetTheme(name string) Theme {
	switch strings.ToLower(name) {
	case "light":
		return LightTheme
	case "cyberpunk":
		return CyberpunkTheme
	case "nord":
		return NordTheme
	case "highcontrast", "high_contrast":
		return HighContrastTheme
	default:
		return DarkTheme
	}
}

type Styles struct {
	Theme        Theme
	Panel        lipgloss.Style
	ActivePanel  lipgloss.Style
	HeaderFg     lipgloss.Style
	MethodGet    lipgloss.Style
	MethodPost   lipgloss.Style
	MethodPut    lipgloss.Style
	MethodPatch  lipgloss.Style
	MethodDelete lipgloss.Style
	MethodOther  lipgloss.Style
	Status2xx    lipgloss.Style
	Status3xx    lipgloss.Style
	Status4xx    lipgloss.Style
	Status5xx    lipgloss.Style
	StatusErr    lipgloss.Style
	Title        lipgloss.Style
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	Label        lipgloss.Style
	Key          lipgloss.Style
	Value        lipgloss.Style
}

func MakeStyles(t Theme) Styles {
	return Styles{
		Theme: t,
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.PanelBorder).
			Padding(0, 1),
		ActivePanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.ActiveBorder).
			Padding(0, 1),
		HeaderFg: lipgloss.NewStyle().
			Foreground(t.HeaderFg).
			Background(t.HeaderBg),
		MethodGet: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.MethodGet),
		MethodPost: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.MethodPost),
		MethodPut: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.MethodPut),
		MethodPatch: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.MethodPatch),
		MethodDelete: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.MethodDelete),
		MethodOther: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Foreground),
		Status2xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Status2xx),
		Status3xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Status3xx),
		Status4xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Status4xx),
		Status5xx: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Status5xx),
		StatusErr: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff5555")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.ActiveBorder),
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.TabActive).
			Underline(true),
		TabInactive: lipgloss.NewStyle().
			Foreground(t.TabInactive),
		Label: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Foreground),
		Key: lipgloss.NewStyle().
			Foreground(t.ActiveBorder),
		Value: lipgloss.NewStyle().
			Foreground(t.Foreground),
	}
}

func (s *Styles) MethodStyle(method string) lipgloss.Style {
	switch strings.ToUpper(method) {
	case "GET":
		return s.MethodGet
	case "POST":
		return s.MethodPost
	case "PUT":
		return s.MethodPut
	case "PATCH":
		return s.MethodPatch
	case "DELETE":
		return s.MethodDelete
	default:
		return s.MethodOther
	}
}

func (s *Styles) StatusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return s.Status2xx
	case code >= 300 && code < 400:
		return s.Status3xx
	case code >= 400 && code < 500:
		return s.Status4xx
	case code >= 500:
		return s.Status5xx
	default:
		return s.StatusErr
	}
}

// PrettyFormatJSON formats and returns pretty JSON
func PrettyFormatJSON(input []byte) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, input, "", "  ")
	if err != nil {
		return string(input), err
	}
	return prettyJSON.String(), nil
}

// HighlightSyntax uses Chroma to highlight JSON or XML for terminal display
func HighlightSyntax(code string, lexerName string, themeName string) string {
	if code == "" {
		return ""
	}
	var buf bytes.Buffer
	// Use terminal256 formatter for CLI
	err := quick.Highlight(&buf, code, lexerName, "terminal256", themeName)
	if err != nil {
		return code
	}
	return buf.String()
}
