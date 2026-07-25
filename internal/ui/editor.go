package ui

import (
	"fmt"
	"strings"

	"vapeu/internal/models"
	"vapeu/internal/theme"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditorSubTab int

const (
	SubTabParams EditorSubTab = iota
	SubTabHeaders
	SubTabAuth
	SubTabBody
	SubTabCookies
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

type EditorPanel struct {
	Request       *models.Request
	styles        theme.Styles
	width         int
	height        int
	focused       bool
	methodIndex   int
	urlInput      textinput.Model
	activeSubTab  EditorSubTab
	bodyInput     textarea.Model
	headersInput  textarea.Model
	paramsInput   textarea.Model
	authInputs    [3]textinput.Model // Basic: User/Pass; Bearer: Token; APIKey: Key/Val
	focusIndex    int                // 0: Method, 1: URL, 2: SubTab bar, 3: Editor Content
}

func NewEditorPanel(req *models.Request, styles theme.Styles) EditorPanel {
	if req == nil {
		req = &models.Request{
			Method: "GET",
			URL:    "https://httpbin.org/get",
			Auth:   models.AuthConfig{Type: models.AuthNone},
			Body:   models.RequestBody{Type: models.BodyJSON},
		}
	}

	urlIn := textinput.New()
	urlIn.Placeholder = "https://api.example.com/v1/resource"
	urlIn.SetValue(req.URL)

	bodyIn := textarea.New()
	bodyIn.Placeholder = "{\n  \"key\": \"value\"\n}"
	bodyIn.SetValue(req.Body.Content)
	bodyIn.SetHeight(6)

	headersIn := textarea.New()
	headersIn.Placeholder = "Header-Name: Header-Value\nUser-Agent: apicli"
	var hLines []string
	for _, h := range req.Headers {
		if h.Enabled {
			hLines = append(hLines, fmt.Sprintf("%s: %s", h.Key, h.Value))
		}
	}
	headersIn.SetValue(strings.Join(hLines, "\n"))
	headersIn.SetHeight(6)

	paramsIn := textarea.New()
	paramsIn.Placeholder = "param1=value1\nparam2=value2"
	var pLines []string
	for _, p := range req.QueryParams {
		if p.Enabled {
			pLines = append(pLines, fmt.Sprintf("%s=%s", p.Key, p.Value))
		}
	}
	paramsIn.SetValue(strings.Join(pLines, "\n"))
	paramsIn.SetHeight(6)

	auth1 := textinput.New()
	auth1.Placeholder = "Username / Bearer Token / Key Name"

	auth2 := textinput.New()
	auth2.Placeholder = "Password / Key Value"

	auth3 := textinput.New()
	auth3.Placeholder = "Header or Query (for API Key)"

	mIdx := 0
	for i, m := range methods {
		if strings.EqualFold(m, req.Method) {
			mIdx = i
			break
		}
	}

	return EditorPanel{
		Request:      req,
		styles:       styles,
		methodIndex:  mIdx,
		urlInput:     urlIn,
		activeSubTab: SubTabBody,
		bodyInput:    bodyIn,
		headersInput: headersIn,
		paramsInput:  paramsIn,
		authInputs:   [3]textinput.Model{auth1, auth2, auth3},
		focusIndex:   1, // focus URL input by default
	}
}

func (p *EditorPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.urlInput.Width = w - 24
	p.bodyInput.SetWidth(w - 6)
	p.headersInput.SetWidth(w - 6)
	p.paramsInput.SetWidth(w - 6)
}

func (p *EditorPanel) SetFocused(focused bool) {
	p.focused = focused
	if focused {
		p.urlInput.Focus()
	} else {
		p.urlInput.Blur()
		p.bodyInput.Blur()
		p.headersInput.Blur()
		p.paramsInput.Blur()
	}
}

func (p EditorPanel) Update(msg tea.Msg) (EditorPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}

		switch msg.String() {
		case "alt+m":
			p.methodIndex = (p.methodIndex + 1) % len(methods)
			p.Request.Method = methods[p.methodIndex]
			return p, nil
		case "alt+1":
			p.activeSubTab = SubTabParams
			return p, nil
		case "alt+2":
			p.activeSubTab = SubTabHeaders
			return p, nil
		case "alt+3":
			p.activeSubTab = SubTabAuth
			return p, nil
		case "alt+4":
			p.activeSubTab = SubTabBody
			return p, nil
		case "down":
			if p.urlInput.Focused() {
				p.urlInput.Blur()
				p.bodyInput.Focus()
			}
		case "up":
			if p.bodyInput.Focused() {
				p.bodyInput.Blur()
				p.urlInput.Focus()
			}
		}
	}

	if p.urlInput.Focused() {
		p.urlInput, cmd = p.urlInput.Update(msg)
		p.Request.URL = p.urlInput.Value()
		return p, cmd
	}

	switch p.activeSubTab {
	case SubTabBody:
		p.bodyInput, cmd = p.bodyInput.Update(msg)
		p.Request.Body.Content = p.bodyInput.Value()
	case SubTabHeaders:
		p.headersInput, cmd = p.headersInput.Update(msg)
		p.parseHeaders()
	case SubTabParams:
		p.paramsInput, cmd = p.paramsInput.Update(msg)
		p.parseParams()
	}

	return p, cmd
}

func (p *EditorPanel) parseHeaders() {
	lines := strings.Split(p.headersInput.Value(), "\n")
	var headers []models.NameValuePair
	for _, l := range lines {
		parts := strings.SplitN(l, ":", 2)
		if len(parts) == 2 {
			headers = append(headers, models.NameValuePair{
				Key:     strings.TrimSpace(parts[0]),
				Value:   strings.TrimSpace(parts[1]),
				Enabled: true,
			})
		}
	}
	p.Request.Headers = headers
}

func (p *EditorPanel) parseParams() {
	lines := strings.Split(p.paramsInput.Value(), "\n")
	var params []models.NameValuePair
	for _, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			params = append(params, models.NameValuePair{
				Key:     strings.TrimSpace(parts[0]),
				Value:   strings.TrimSpace(parts[1]),
				Enabled: true,
			})
		}
	}
	p.Request.QueryParams = params
}

func (p EditorPanel) View() string {
	style := p.styles.Panel
	if p.focused {
		style = p.styles.ActivePanel
	}

	methodStr := p.styles.MethodStyle(methods[p.methodIndex]).Render(fmt.Sprintf(" [%s] ", methods[p.methodIndex]))
	urlView := p.urlInput.View()
	sendBtn := lipgloss.NewStyle().Background(lipgloss.Color("#cba6f7")).Foreground(lipgloss.Color("#11111b")).Bold(true).Render(" Ctrl+R Send ")

	topBar := lipgloss.JoinHorizontal(lipgloss.Center, methodStr, " ", urlView, " ", sendBtn)

	// Sub-tabs
	subTabs := []string{"Alt+1 Params", "Alt+2 Headers", "Alt+3 Auth", "Alt+4 Body", "Cookies"}
	var renderedTabs []string
	for i, t := range subTabs {
		if EditorSubTab(i) == p.activeSubTab {
			renderedTabs = append(renderedTabs, p.styles.TabActive.Render(fmt.Sprintf("[%s]", t)))
		} else {
			renderedTabs = append(renderedTabs, p.styles.TabInactive.Render(t))
		}
	}
	tabBar := strings.Join(renderedTabs, "  ")

	// Active tab content
	var tabContent string
	switch p.activeSubTab {
	case SubTabParams:
		tabContent = "Query Parameters (key=value per line):\n" + p.paramsInput.View()
	case SubTabHeaders:
		tabContent = "Headers (Header-Name: Value per line):\n" + p.headersInput.View()
	case SubTabAuth:
		authType := p.Request.Auth.Type
		if authType == "" {
			authType = models.AuthNone
		}
		tabContent = fmt.Sprintf("Authentication Type: %s (Bearer / Basic / APIKey)\n\nBearer Token / Auth details configured in Request object.", authType)
	case SubTabBody:
		tabContent = "Request Body (JSON / Raw):\n" + p.bodyInput.View()
	case SubTabCookies:
		tabContent = "Cookies (Managed automatically or via Headers)."
	}

	content := lipgloss.JoinVertical(lipgloss.Left, topBar, "\n", tabBar, "\n", tabContent)

	return style.
		Width(p.width - 2).
		Height(p.height - 2).
		Render(content)
}

func (p *EditorPanel) LoadRequest(req *models.Request) {
	if req == nil {
		return
	}
	p.Request = req
	p.urlInput.SetValue(req.URL)
	p.bodyInput.SetValue(req.Body.Content)

	for i, m := range methods {
		if strings.EqualFold(m, req.Method) {
			p.methodIndex = i
			break
		}
	}

	var hLines []string
	for _, h := range req.Headers {
		if h.Enabled {
			hLines = append(hLines, fmt.Sprintf("%s: %s", h.Key, h.Value))
		}
	}
	p.headersInput.SetValue(strings.Join(hLines, "\n"))
}
