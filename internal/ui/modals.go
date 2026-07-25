package ui

import (
	"fmt"
	"strings"

	"vapeu/internal/impexp"
	"vapeu/internal/models"
	"vapeu/internal/theme"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalSearch
	ModalHistory
	ModalEnv
	ModalImport
	ModalHelp
)

type Modals struct {
	ActiveModal  ModalType
	styles       theme.Styles
	searchInput  textinput.Model
	importInput  textarea.Model
	historyItems []models.HistoryItem
	envItems     []models.Environment
	selectedIdx  int
}

func NewModals(styles theme.Styles) Modals {
	sIn := textinput.New()
	sIn.Placeholder = "Search by request name, URL or method..."

	impIn := textarea.New()
	impIn.Placeholder = "Paste cURL command or OpenAPI YAML/JSON here..."
	impIn.SetHeight(8)

	return Modals{
		ActiveModal: ModalNone,
		styles:      styles,
		searchInput: sIn,
		importInput: impIn,
	}
}

func (m *Modals) Show(t ModalType) {
	m.ActiveModal = t
	m.selectedIdx = 0
	if t == ModalSearch {
		m.searchInput.Focus()
	} else if t == ModalImport {
		m.importInput.Focus()
	}
}

func (m *Modals) Hide() {
	m.ActiveModal = ModalNone
	m.searchInput.Blur()
	m.importInput.Blur()
}

func (m Modals) Update(msg tea.Msg) (Modals, tea.Cmd, *models.Request) {
	if m.ActiveModal == ModalNone {
		return m, nil, nil
	}

	var cmd tea.Cmd
	var reqToLoad *models.Request

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, nil, nil
		}
	}

	switch m.ActiveModal {
	case ModalSearch:
		m.searchInput, cmd = m.searchInput.Update(msg)
	case ModalImport:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "ctrl+s" || msg.String() == "enter" {
				val := strings.TrimSpace(m.importInput.Value())
				if strings.HasPrefix(val, "curl") {
					if req, err := impexp.ParseCurl(val); err == nil {
						reqToLoad = req
						m.Hide()
					}
				}
			}
		}
		m.importInput, cmd = m.importInput.Update(msg)
	case ModalHistory:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			case "down":
				if m.selectedIdx < len(m.historyItems)-1 {
					m.selectedIdx++
				}
			case "enter":
				if m.selectedIdx >= 0 && m.selectedIdx < len(m.historyItems) {
					req := m.historyItems[m.selectedIdx].Request
					reqToLoad = &req
					m.Hide()
				}
			}
		}
	}

	return m, cmd, reqToLoad
}

func (m Modals) View(w, h int) string {
	if m.ActiveModal == ModalNone {
		return ""
	}

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(m.styles.Theme.ActiveBorder).
		Background(m.styles.Theme.HeaderBg).
		Padding(1, 2).
		Width(w - 20)

	var title, body string

	switch m.ActiveModal {
	case ModalSearch:
		title = "Search Requests (Press ESC to close)"
		body = m.searchInput.View()
	case ModalImport:
		title = "Import Request (cURL / OpenAPI) (Press Enter to import, ESC to cancel)"
		body = m.importInput.View()
	case ModalHistory:
		title = "Request History (Up/Down to navigate, Enter to load, ESC to close)"
		var sb strings.Builder
		if len(m.historyItems) == 0 {
			sb.WriteString("No request history recorded yet.")
		} else {
			for i, item := range m.historyItems {
				prefix := "  "
				if i == m.selectedIdx {
					prefix = "> "
				}
				sb.WriteString(fmt.Sprintf("%s[%s] %s (%d ms)\n", prefix, item.Request.Method, item.Request.URL, item.DurationMs))
			}
		}
		body = sb.String()
	case ModalHelp:
		title = "Keyboard Shortcuts Reference"
		body = `
  Ctrl+N       New Request
  Ctrl+S       Save Request
  Ctrl+R       Send Request
  Ctrl+F       Search Requests
  Ctrl+H       Request History
  Ctrl+I       Import cURL / OpenAPI
  Ctrl+T       New Tab
  Ctrl+W       Close Tab
  Ctrl+Q       Quit App
  Tab          Next Panel Focus
  Shift+Tab    Previous Panel Focus
  Alt+1..5     Switch Sub-Tabs (Params, Headers, Auth, Body, Cookies)
  ESC          Close Modal / Cancel
`
	}

	content := fmt.Sprintf("%s\n\n%s", m.styles.Title.Render(title), body)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, modalBox.Render(content))
}
