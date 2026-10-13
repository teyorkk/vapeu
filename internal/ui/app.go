package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"vapeu/internal/api"
	"vapeu/internal/models"
	"vapeu/internal/storage"
	"vapeu/internal/theme"
	"vapeu/internal/variables"
	"github.com/google/uuid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PanelFocus int

const (
	FocusSidebar PanelFocus = iota
	FocusEditor
	FocusResponse
)

type tabItem struct {
	Title   string
	Request *models.Request
}

type splashTimeoutMsg struct{}

type AppModel struct {
	storage      *storage.Storage
	config       models.Config
	styles       theme.Styles
	client       *api.Client
	collections  []models.Collection
	environments []models.Environment
	history      []models.HistoryItem
	activeEnv    *models.Environment

	showSplash        bool
	showSidebar       bool
	activeFocus       PanelFocus
	width             int
	height            int
	requestCancelFunc context.CancelFunc

	// Panels
	sidebarPanel  SidebarPanel
	editorPanel   EditorPanel
	responsePanel ResponsePanel
	modals        Modals

	// Tabs
	tabs      []tabItem
	activeTab int
}

type responseMsg struct {
	resp *models.Response
	err  error
}

func NewAppModel(st *storage.Storage) AppModel {
	cfg, _ := st.LoadConfig()
	th := theme.GetTheme(cfg.Theme)
	styles := theme.MakeStyles(th)

	cols, _ := st.LoadCollections()
	envs, _ := st.LoadEnvironments()
	hist, _ := st.LoadHistory()

	if len(cols) == 0 {
		demoCol := models.Collection{
			ID:   uuid.New().String(),
			Name: "Demo API",
			Nodes: []*models.CollectionNode{
				{
					ID:   uuid.New().String(),
					Name: "JSONPlaceholder Todo",
					Kind: models.NodeRequest,
					Request: &models.Request{
						ID:     uuid.New().String(),
						Name:   "JSONPlaceholder Todo",
						Method: "GET",
						URL:    "https://jsonplaceholder.typicode.com/todos/1",
						Auth:   models.AuthConfig{Type: models.AuthNone},
						Body:   models.RequestBody{Type: models.BodyNone},
					},
				},
				{
					ID:   uuid.New().String(),
					Name: "JSONPlaceholder Create Post",
					Kind: models.NodeRequest,
					Request: &models.Request{
						ID:     uuid.New().String(),
						Name:   "JSONPlaceholder Create Post",
						Method: "POST",
						URL:    "https://jsonplaceholder.typicode.com/posts",
						Auth:   models.AuthConfig{Type: models.AuthNone},
						Body: models.RequestBody{
							Type:    models.BodyJSON,
							Content: "{\n  \"title\": \"foo\",\n  \"body\": \"bar\",\n  \"userId\": 1\n}",
						},
					},
				},
			},
		}
		_ = st.SaveCollection(demoCol)
		cols = []models.Collection{demoCol}
	}

	client := api.NewClient(api.ClientOptions{
		TimeoutSec:  cfg.DefaultTimeoutSec,
		InsecureSSL: !cfg.SSLVerification,
		ProxyURL:    cfg.ProxyURL,
	})

	firstReq := demoRequest()
	if len(cols) > 0 && len(cols[0].Nodes) > 0 && cols[0].Nodes[0].Request != nil {
		firstReq = cols[0].Nodes[0].Request
	}

	sidebarPanel := NewSidebarPanel(styles, cols)
	editorPanel := NewEditorPanel(firstReq, styles)
	respPanel := NewResponsePanel(styles)
	modals := NewModals(styles)

	tabs := []tabItem{
		{Title: firstReq.Name, Request: firstReq},
	}

	app := AppModel{
		storage:       st,
		config:        cfg,
		styles:        styles,
		client:        client,
		collections:   cols,
		environments:  envs,
		history:       hist,
		showSplash:    true,
		showSidebar:   true,
		activeFocus:   FocusEditor,
		sidebarPanel:  sidebarPanel,
		editorPanel:   editorPanel,
		responsePanel: respPanel,
		modals:        modals,
		tabs:          tabs,
		activeTab:     0,
	}

	app.updateFocus()
	return app
}

func demoRequest() *models.Request {
	return &models.Request{
		ID:     uuid.New().String(),
		Name:   "JSONPlaceholder Todo",
		Method: "GET",
		URL:    "https://jsonplaceholder.typicode.com/todos/1",
		Auth:   models.AuthConfig{Type: models.AuthNone},
		Body:   models.RequestBody{Type: models.BodyNone},
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Tick(1200*time.Millisecond, func(t time.Time) tea.Msg {
		return splashTimeoutMsg{}
	})
}

func (m *AppModel) updateFocus() {
	m.sidebarPanel.SetFocused(m.activeFocus == FocusSidebar && m.modals.ActiveModal == ModalNone && !m.showSplash)
	m.editorPanel.SetFocused(m.activeFocus == FocusEditor && m.modals.ActiveModal == ModalNone && !m.showSplash)
	m.responsePanel.SetFocused(m.activeFocus == FocusResponse && m.modals.ActiveModal == ModalNone && !m.showSplash)
}

func (m *AppModel) getAllRequests() []*models.Request {
	var list []*models.Request
	seen := make(map[string]bool)

	for _, col := range m.collections {
		for _, node := range col.Nodes {
			if node.Request != nil && !seen[node.Request.ID] {
				seen[node.Request.ID] = true
				list = append(list, node.Request)
			}
		}
	}

	for _, item := range m.history {
		if !seen[item.Request.ID] {
			seen[item.Request.ID] = true
			req := item.Request
			list = append(list, &req)
		}
	}

	return list
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case splashTimeoutMsg:
		m.showSplash = false
		m.updateFocus()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateLayout()
		return m, nil

	case responseMsg:
		m.responsePanel.SetLoading(false)
		m.requestCancelFunc = nil
		if msg.resp != nil {
			m.responsePanel.SetResponse(msg.resp)
			histItem := models.HistoryItem{
				ID:         uuid.New().String(),
				Request:    *m.editorPanel.Request,
				Response:   msg.resp,
				DurationMs: msg.resp.ResponseTimeMs,
				StatusCode: msg.resp.StatusCode,
			}
			m.history = append([]models.HistoryItem{histItem}, m.history...)
			m.modals.historyItems = m.history
			_ = m.storage.AddHistoryItem(histItem)
		}
		return m, nil

	case tea.KeyMsg:
		if m.showSplash {
			m.showSplash = false
			m.updateFocus()
			return m, nil
		}

		if m.modals.ActiveModal != ModalNone {
			var cmd tea.Cmd
			var reqToLoad *models.Request
			var savedName string
			var newCol *models.Collection
			m.modals, cmd, reqToLoad, savedName, newCol = m.modals.Update(msg)

			if reqToLoad != nil {
				m.loadRequest(reqToLoad)
			}
			if savedName != "" {
				m.saveCurrentRequest(savedName)
			}
			if newCol != nil {
				m.collections = append(m.collections, *newCol)
				_ = m.storage.SaveCollection(*newCol)
				m.sidebarPanel.SetCollections(m.collections)
			}
			return m, cmd
		}

		switch msg.String() {
		case "alt+s":
			m.showSidebar = !m.showSidebar
			if !m.showSidebar && m.activeFocus == FocusSidebar {
				m.activeFocus = FocusEditor
			}
			m.recalculateLayout()
			m.updateFocus()
			return m, nil
		case "alt+k":
			m.cycleTheme()
			return m, nil
		case "ctrl+q":
			return m, tea.Quit
		case "ctrl+x":
			m.cancelRunningRequest()
			return m, nil
		case "ctrl+s":
			m.modals.ShowSave(m.editorPanel.Request.Name)
			m.updateFocus()
			return m, nil
		case "alt+ ", "alt+space", "ctrl+f":
			m.modals.ShowTelescope(m.getAllRequests())
			m.updateFocus()
			return m, nil
		case "tab", "shift+tab":
			if m.showSidebar {
				if m.activeFocus == FocusSidebar {
					m.activeFocus = FocusEditor
				} else if m.activeFocus == FocusEditor {
					m.activeFocus = FocusResponse
				} else {
					m.activeFocus = FocusSidebar
				}
			} else {
				if m.activeFocus == FocusEditor {
					m.activeFocus = FocusResponse
				} else {
					m.activeFocus = FocusEditor
				}
			}
			m.updateFocus()
			return m, nil
		case "ctrl+1":
			if m.showSidebar {
				m.activeFocus = FocusSidebar
			} else {
				m.activeFocus = FocusEditor
			}
			m.updateFocus()
			return m, nil
		case "ctrl+2":
			m.activeFocus = FocusEditor
			m.updateFocus()
			return m, nil
		case "ctrl+3":
			m.activeFocus = FocusResponse
			m.updateFocus()
			return m, nil
		case "alt+left", "alt+[":
			if len(m.tabs) > 1 {
				m.switchTab((m.activeTab - 1 + len(m.tabs)) % len(m.tabs))
			}
			return m, nil
		case "alt+right", "alt+]":
			if len(m.tabs) > 1 {
				m.switchTab((m.activeTab + 1) % len(m.tabs))
			}
			return m, nil
		case "ctrl+n", "ctrl+t":
			newReq := demoRequest()
			m.tabs = append(m.tabs, tabItem{Title: "New Request", Request: newReq})
			m.activeTab = len(m.tabs) - 1
			m.editorPanel.LoadRequest(newReq)
			return m, nil
		case "ctrl+w":
			if len(m.tabs) > 1 {
				m.tabs = append(m.tabs[:m.activeTab], m.tabs[m.activeTab+1:]...)
				if m.activeTab >= len(m.tabs) {
					m.activeTab = len(m.tabs) - 1
				}
				m.editorPanel.LoadRequest(m.tabs[m.activeTab].Request)
			}
			return m, nil
		case "ctrl+r", "ctrl+enter":
			m.responsePanel.SetLoading(true)
			req := m.editorPanel.Request
			resolver := variables.NewResolver(nil, nil, nil, nil)
			if m.activeEnv != nil {
				resolver = variables.NewResolver(m.activeEnv.Variables, nil, nil, nil)
			}

			ctx, cancel := context.WithCancel(context.Background())
			m.requestCancelFunc = cancel

			return m, func() tea.Msg {
				resp, err := m.client.ExecuteWithContext(ctx, req, resolver)
				return responseMsg{resp: resp, err: err}
			}
		case "ctrl+h":
			m.modals.historyItems = m.history
			m.modals.Show(ModalHistory)
			m.updateFocus()
			return m, nil
		case "ctrl+i", "alt+i":
			m.modals.Show(ModalImport)
			m.updateFocus()
			return m, nil
		case "?":
			m.modals.Show(ModalHelp)
			m.updateFocus()
			return m, nil
		}
	}

	// Route update to focused panel
	switch m.activeFocus {
	case FocusSidebar:
		var cmd tea.Cmd
		var reqToLoad *models.Request
		var triggerNewCol, triggerNewReq bool
		m.sidebarPanel, cmd, reqToLoad, triggerNewCol, triggerNewReq = m.sidebarPanel.Update(msg)
		cmds = append(cmds, cmd)

		if reqToLoad != nil {
			m.loadRequest(reqToLoad)
		}
		if triggerNewCol {
			m.modals.ShowCreateCollection()
			m.updateFocus()
		}
		if triggerNewReq {
			m.createRequestUnderCurrentCollection()
		}

	case FocusEditor:
		var cmd tea.Cmd
		m.editorPanel, cmd = m.editorPanel.Update(msg)
		cmds = append(cmds, cmd)
		if m.activeTab < len(m.tabs) {
			m.tabs[m.activeTab].Request = m.editorPanel.Request
			if m.editorPanel.Request.Name != "" {
				m.tabs[m.activeTab].Title = m.editorPanel.Request.Name
			}
		}
	case FocusResponse:
		var cmd tea.Cmd
		m.responsePanel, cmd = m.responsePanel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) cancelRunningRequest() {
	if m.requestCancelFunc != nil {
		m.requestCancelFunc()
		m.requestCancelFunc = nil
	}
	m.responsePanel.SetLoading(false)
	m.responsePanel.SetResponse(&models.Response{
		StatusCode: 0,
		Error:      "Request cancelled by user (Ctrl+X)",
		Timestamp:  time.Now().Format(time.RFC3339),
	})
}

func (m *AppModel) saveCurrentRequest(name string) {
	if name == "" {
		return
	}

	req := m.editorPanel.Request
	req.Name = name
	m.tabs[m.activeTab].Title = name

	if len(m.collections) > 0 {
		col := &m.collections[0]
		found := false
		for _, node := range col.Nodes {
			if node.Request != nil && node.Request.ID == req.ID {
				node.Name = name
				node.Request = req
				found = true
				break
			}
		}
		if !found {
			col.Nodes = append(col.Nodes, &models.CollectionNode{
				ID:      uuid.New().String(),
				Name:    name,
				Kind:    models.NodeRequest,
				Request: req,
			})
		}
		_ = m.storage.SaveCollection(*col)
		m.sidebarPanel.SetCollections(m.collections)
	}
}

func (m *AppModel) createRequestUnderCurrentCollection() {
	if len(m.collections) == 0 {
		return
	}

	targetColIdx := 0
	item := m.sidebarPanel.SelectedItem()
	if item != nil && item.CollectionIdx >= 0 && item.CollectionIdx < len(m.collections) {
		targetColIdx = item.CollectionIdx
	}

	col := &m.collections[targetColIdx]

	newURL := col.BaseURL
	if newURL != "" && !strings.HasSuffix(newURL, "/") {
		newURL += "/"
	}

	newReq := &models.Request{
		ID:     uuid.New().String(),
		Name:   fmt.Sprintf("New Request %d", len(col.Nodes)+1),
		Method: "GET",
		URL:    newURL,
		Auth:   models.AuthConfig{Type: models.AuthNone},
		Body:   models.RequestBody{Type: models.BodyNone},
	}

	col.Nodes = append(col.Nodes, &models.CollectionNode{
		ID:      uuid.New().String(),
		Name:    newReq.Name,
		Kind:    models.NodeRequest,
		Request: newReq,
	})

	_ = m.storage.SaveCollection(*col)
	m.sidebarPanel.SetCollections(m.collections)

	// Also load into editor
	m.tabs = append(m.tabs, tabItem{Title: newReq.Name, Request: newReq})
	m.activeTab = len(m.tabs) - 1
	m.editorPanel.LoadRequest(newReq)
	m.activeFocus = FocusEditor
	m.updateFocus()
}

func (m *AppModel) cycleTheme() {
	curr := strings.ToLower(m.config.Theme)
	idx := 0
	for i, t := range theme.AvailableThemes {
		if t == curr {
			idx = (i + 1) % len(theme.AvailableThemes)
			break
		}
	}
	nextTheme := theme.AvailableThemes[idx]
	m.config.Theme = nextTheme
	_ = m.storage.SaveConfig(m.config)

	th := theme.GetTheme(nextTheme)
	m.styles = theme.MakeStyles(th)
	m.sidebarPanel.styles = m.styles
	m.editorPanel.styles = m.styles
	m.responsePanel.styles = m.styles
	m.modals.styles = m.styles
}

func (m *AppModel) switchTab(index int) {
	if index >= 0 && index < len(m.tabs) {
		m.activeTab = index
		m.editorPanel.LoadRequest(m.tabs[index].Request)
	}
}

func (m *AppModel) loadRequest(req *models.Request) {
	m.editorPanel.LoadRequest(req)
	m.tabs[m.activeTab].Request = req
	m.tabs[m.activeTab].Title = req.Name
	m.activeFocus = FocusEditor
	m.updateFocus()
}

func (m *AppModel) recalculateLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	// 1 line header
	availHeight := m.height - 1
	if availHeight < 10 {
		availHeight = 10
	}

	topHeight := (availHeight * 50) / 100
	bottomHeight := availHeight - topHeight

	mainWidth := m.width
	if m.showSidebar {
		sidebarW := 28
		if m.width < 70 {
			sidebarW = 20
		}
		mainWidth = m.width - sidebarW
		if mainWidth < 20 {
			mainWidth = 20
		}
		m.sidebarPanel.SetSize(sidebarW, availHeight)
	}

	m.editorPanel.SetSize(mainWidth, topHeight)
	m.responsePanel.SetSize(mainWidth, bottomHeight)
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing TUI..."
	}

	if m.showSplash {
		return m.renderSplash()
	}

	// Header with VAPEU branding & Telescope prompt hint
	brandBadge := lipgloss.NewStyle().
		Background(lipgloss.Color("#cba6f7")).
		Foreground(lipgloss.Color("#11111b")).
		Bold(true).
		Render(" VAPEU ")

	var tabTitles []string
	for i, t := range m.tabs {
		title := t.Title
		if title == "" {
			title = "Request"
		}
		if i == m.activeTab {
			tabTitles = append(tabTitles, m.styles.TabActive.Render(fmt.Sprintf("▶ [%d: %s] ◀", i+1, title)))
		} else {
			tabTitles = append(tabTitles, m.styles.TabInactive.Render(fmt.Sprintf(" %d: %s ", i+1, title)))
		}
	}

	tabBar := strings.Join(tabTitles, " | ")
	newTabBtn := lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Render(" [+ Ctrl+T] ")
	sidebarHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render(" [Alt+S Sidebar] ")
	telescopeHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Render(" [Alt+Space Telescope] ")

	headerView := m.styles.HeaderFg.Render(brandBadge + " " + tabBar + newTabBtn + sidebarHint + telescopeHint + " | Press '?' for Help ")

	// Render Editor & Response Panels
	editorView := m.editorPanel.View()
	responseView := m.responsePanel.View()
	rightStack := lipgloss.JoinVertical(lipgloss.Left, editorView, responseView)

	var bodyView string
	if m.showSidebar {
		sidebarView := m.sidebarPanel.View()
		bodyView = lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, rightStack)
	} else {
		bodyView = rightStack
	}

	mainLayout := lipgloss.JoinVertical(lipgloss.Left, headerView, bodyView)

	// Ensure total layout height NEVER overflows terminal height
	lines := strings.Split(mainLayout, "\n")
	if len(lines) > m.height {
		mainLayout = strings.Join(lines[:m.height], "\n")
	}

	// Overlay Modals if active
	if m.modals.ActiveModal != ModalNone {
		return m.modals.View(m.width, m.height)
	}

	return mainLayout
}

func (m AppModel) renderSplash() string {
	banner := `
__   __   _   ___  _____ _   _ 
\ \ / /  /_\ | _ \/ __/  | | | |
 \ V /  / _ \|  _/| _|   | |_| |
  \_/  /_/ \_\_|  |___|   \___/ 
`

	logoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#cba6f7"))

	subText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6adc8")).
		Render("v a p e u  --  Terminal API Client\n\nInitializing client interface...")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#cba6f7")).
		Padding(2, 4).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(lipgloss.Center, logoStyle.Render(banner), "\n", subText)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box.Render(content))
}
