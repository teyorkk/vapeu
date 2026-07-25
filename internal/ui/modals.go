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
	ModalSave
)

type Modals struct {
	ActiveModal  ModalType
	styles       theme.Styles
	searchInput  textinput.Model
	importInput  textarea.Model
	saveInput    textinput.Model
	historyItems []models.HistoryItem
	allRequests  []*models.Request
	envItems     []models.Environment
	selectedIdx  int
}

func NewModals(styles theme.Styles) Modals {
	sIn := textinput.New()
	sIn.Placeholder = "Type to search requests by name, URL, or method..."

	impIn := textarea.New()
	impIn.Placeholder = "Paste cURL command or OpenAPI YAML/JSON here..."
	impIn.SetHeight(8)

	saveIn := textinput.New()
	saveIn.Placeholder = "Enter request name..."

	return Modals{
		ActiveModal: ModalNone,
		styles:      styles,
		searchInput: sIn,
		importInput: impIn,
		saveInput:   saveIn,
	}
}

func (m *Modals) Show(t ModalType) {
	m.ActiveModal = t
	m.selectedIdx = 0
	if t == ModalSearch {
		m.searchInput.SetValue("")
		m.searchInput.Focus()
	} else if t == ModalImport {
		m.importInput.Focus()
	} else if t == ModalSave {
		m.saveInput.Focus()
	}
}

func (m *Modals) ShowTelescope(reqs []*models.Request) {
	m.allRequests = reqs
	m.Show(ModalSearch)
}

func (m *Modals) ShowSave(currentName string) {
	m.saveInput.SetValue(currentName)
	m.Show(ModalSave)
}

func (m *Modals) Hide() {
	m.ActiveModal = ModalNone
	m.searchInput.Blur()
	m.importInput.Blur()
	m.saveInput.Blur()
}

func (m Modals) filteredRequests() []*models.Request {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		return m.allRequests
	}

	var results []*models.Request
	for _, req := range m.allRequests {
		if req == nil {
			continue
		}
		if strings.Contains(strings.ToLower(req.Name), query) ||
			strings.Contains(strings.ToLower(req.URL), query) ||
			strings.Contains(strings.ToLower(req.Method), query) {
			results = append(results, req)
		}
	}
	return results
}

func (m Modals) Update(msg tea.Msg) (Modals, tea.Cmd, *models.Request, string) {
	if m.ActiveModal == ModalNone {
		return m, nil, nil, ""
	}

	var cmd tea.Cmd
	var reqToLoad *models.Request
	var savedName string

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, nil, nil, ""
		}
	}

	switch m.ActiveModal {
	case ModalSearch:
		filtered := m.filteredRequests()
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
				return m, nil, nil, ""
			case "down":
				if m.selectedIdx < len(filtered)-1 {
					m.selectedIdx++
				}
				return m, nil, nil, ""
			case "enter":
				if m.selectedIdx >= 0 && m.selectedIdx < len(filtered) {
					reqToLoad = filtered[m.selectedIdx]
					m.Hide()
					return m, nil, reqToLoad, ""
				}
			}
		}
		m.searchInput, cmd = m.searchInput.Update(msg)

	case ModalSave:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				name := strings.TrimSpace(m.saveInput.Value())
				if name != "" {
					savedName = name
					m.Hide()
				}
			}
		}
		m.saveInput, cmd = m.saveInput.Update(msg)

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

	return m, cmd, reqToLoad, savedName
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
		Width(w - 16)

	var title, body string

	switch m.ActiveModal {
	case ModalSearch:
		title = "🔭 Telescope Request Finder (Alt+Space / Ctrl+F) - Press ESC to exit"
		filtered := m.filteredRequests()

		var sb strings.Builder
		sb.WriteString(m.searchInput.View() + "\n\n")

		if len(filtered) == 0 {
			sb.WriteString("  No matching requests found.")
		} else {
			maxDisplay := 10
			if len(filtered) < maxDisplay {
				maxDisplay = len(filtered)
			}
			for i := 0; i < maxDisplay; i++ {
				req := filtered[i]
				prefix := "  "
				if i == m.selectedIdx {
					prefix = "> "
				}
				methodBadge := m.styles.MethodStyle(req.Method).Render(fmt.Sprintf("[%s]", req.Method))
				sb.WriteString(fmt.Sprintf("%s%s  %-20s %s\n", prefix, methodBadge, req.Name, m.styles.Value.Render(req.URL)))
			}
		}
		body = sb.String()

	case ModalSave:
		title = "Save Request (Enter Name and press Enter)"
		body = m.saveInput.View()

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
				methodBadge := m.styles.MethodStyle(item.Request.Method).Render(fmt.Sprintf("[%s]", item.Request.Method))
				sb.WriteString(fmt.Sprintf("%s%s  %s (%d ms)\n", prefix, methodBadge, item.Request.URL, item.DurationMs))
			}
		}
		body = sb.String()

	case ModalHelp:
		title = "Keyboard Shortcuts Reference"
		body = `
  Alt+Space    🔭 Telescope Request Finder (List & Search Requests)
  Alt+M        Toggle HTTP Method (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS)
  Alt+K        🎨 Cycle UI Color Theme (Dark, Light, Cyberpunk, Nord, HighContrast)
  Ctrl+N / T   New Request / New Tab
  Alt+Left / Right (or Alt+[ / ]) Switch Open Request Tabs
  Ctrl+S       Save Request (Prompts for Name)
  Ctrl+R       Send Request
  Ctrl+X       Stop / Cancel Running Request
  Ctrl+H       Request History
  Alt+I / Ctrl+I Import cURL / OpenAPI
  Ctrl+W       Close Tab
  Ctrl+Q       Quit App
  Tab / Shift  Switch Panel Focus (Editor <-> Response)
  Ctrl+1 / 2   Focus Request Editor / Focus Response Viewer
  Alt+1..5     Switch Sub-Tabs (Params, Headers, Auth, Body, Cookies)
  ESC / Alt+U  Focus URL Bar (from sub-tab text area) / Close Modal
`
	}

	content := fmt.Sprintf("%s\n\n%s", m.styles.Title.Render(title), body)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, modalBox.Render(content))
}
