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
	FocusCollections PanelFocus = iota
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
	activeFocus       PanelFocus
	width             int
	height            int
	requestCancelFunc context.CancelFunc

	// Panels
	collectionsPanel CollectionsPanel
	editorPanel      EditorPanel
	responsePanel    ResponsePanel
	modals           Modals

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
					Name: "HTTPBin Get",
					Kind: models.NodeRequest,
					Request: &models.Request{
						ID:     uuid.New().String(),
						Name:   "HTTPBin Get",
						Method: "GET",
						URL:    "https://httpbin.org/get",
						Auth:   models.AuthConfig{Type: models.AuthNone},
						Body:   models.RequestBody{Type: models.BodyNone},
					},
				},
				{
					ID:   uuid.New().String(),
					Name: "HTTPBin Post JSON",
					Kind: models.NodeRequest,
					Request: &models.Request{
						ID:     uuid.New().String(),
						Name:   "HTTPBin Post JSON",
						Method: "POST",
						URL:    "https://httpbin.org/post",
						Auth:   models.AuthConfig{Type: models.AuthNone},
						Body: models.RequestBody{
							Type:    models.BodyJSON,
							Content: "{\n  \"hello\": \"world\",\n  \"user\": \"{{USER_ID}}\"\n}",
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

	collPanel := NewCollectionsPanel(cols, styles)
	firstReq := demoRequest()
	if len(cols) > 0 && len(cols[0].Nodes) > 0 && cols[0].Nodes[0].Request != nil {
		firstReq = cols[0].Nodes[0].Request
	}

	editorPanel := NewEditorPanel(firstReq, styles)
	respPanel := NewResponsePanel(styles)
	modals := NewModals(styles)

	tabs := []tabItem{
		{Title: firstReq.Name, Request: firstReq},
	}

	app := AppModel{
		storage:          st,
		config:           cfg,
		styles:           styles,
		client:           client,
		collections:      cols,
		environments:     envs,
		history:          hist,
		showSplash:       true,
		activeFocus:      FocusEditor,
		collectionsPanel: collPanel,
		editorPanel:      editorPanel,
		responsePanel:    respPanel,
		modals:           modals,
		tabs:             tabs,
		activeTab:        0,
	}

	app.updateFocus()
	return app
}

func demoRequest() *models.Request {
	return &models.Request{
		ID:     uuid.New().String(),
		Name:   "New Request",
		Method: "GET",
		URL:    "https://httpbin.org/get",
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
	m.collectionsPanel.SetFocused(m.activeFocus == FocusCollections && m.modals.ActiveModal == ModalNone && !m.showSplash)
	m.editorPanel.SetFocused(m.activeFocus == FocusEditor && m.modals.ActiveModal == ModalNone && !m.showSplash)
	m.responsePanel.SetFocused(m.activeFocus == FocusResponse && m.modals.ActiveModal == ModalNone && !m.showSplash)
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
			m.modals, cmd, reqToLoad, savedName = m.modals.Update(msg)

			if reqToLoad != nil {
				m.loadRequest(reqToLoad)
			}
			if savedName != "" {
				m.saveCurrentRequest(savedName)
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+q":
			return m, tea.Quit
		case "ctrl+x":
			m.cancelRunningRequest()
			return m, nil
		case "ctrl+s":
			m.modals.ShowSave(m.editorPanel.Request.Name)
			m.updateFocus()
			return m, nil
		case "tab":
			m.activeFocus = (m.activeFocus + 1) % 3
			m.updateFocus()
			return m, nil
		case "shift+tab":
			m.activeFocus = (m.activeFocus + 2) % 3
			m.updateFocus()
			return m, nil
		case "ctrl+1":
			m.activeFocus = FocusCollections
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
		case "ctrl+f":
			m.modals.Show(ModalSearch)
			m.updateFocus()
			return m, nil
		case "ctrl+h":
			m.modals.historyItems = m.history
			m.modals.Show(ModalHistory)
			m.updateFocus()
			return m, nil
		case "ctrl+i":
			m.modals.Show(ModalImport)
			m.updateFocus()
			return m, nil
		case "?":
			m.modals.Show(ModalHelp)
			m.updateFocus()
			return m, nil
		case "enter":
			if m.activeFocus == FocusCollections {
				selectedReq := m.collectionsPanel.SelectedRequest()
				if selectedReq != nil {
					m.loadRequest(selectedReq)
				}
			}
		}
	}

	// Route update to focused panel
	switch m.activeFocus {
	case FocusCollections:
		var cmd tea.Cmd
		m.collectionsPanel, cmd = m.collectionsPanel.Update(msg)
		cmds = append(cmds, cmd)
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

	// Find or update in collection
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
		m.collectionsPanel.Refresh(m.collections)
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

	topHeight := (m.height * 55) / 100
	bottomHeight := m.height - topHeight - 2

	leftWidth := (m.width * 30) / 100
	rightWidth := m.width - leftWidth

	m.collectionsPanel.SetSize(leftWidth, topHeight)
	m.editorPanel.SetSize(rightWidth, topHeight)
	m.responsePanel.SetSize(m.width, bottomHeight)
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing TUI..."
	}

	if m.showSplash {
		return m.renderSplash()
	}

	// Render Header with VAPEU branding & clear Tab Indicators
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
	tabCountInfo := fmt.Sprintf("(%d Tabs)", len(m.tabs))

	headerView := m.styles.HeaderFg.Render(brandBadge + " " + tabBar + newTabBtn + " " + tabCountInfo + " | Press '?' for Help ")

	// Render 3 Panels
	collView := m.collectionsPanel.View()
	editorView := m.editorPanel.View()
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, collView, editorView)

	responseView := m.responsePanel.View()

	mainLayout := lipgloss.JoinVertical(lipgloss.Left, headerView, topSection, responseView)

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
