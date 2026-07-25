package ui

import (
	"fmt"
	"strings"

	"vapeu/internal/models"
	"vapeu/internal/theme"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResponseSubTab int

const (
	RespTabBody ResponseSubTab = iota
	RespTabHeaders
	RespTabCookies
)

type ResponsePanel struct {
	Response     *models.Response
	styles       theme.Styles
	width        int
	height       int
	focused      bool
	viewport     viewport.Model
	activeSubTab ResponseSubTab
	isPretty     bool
	isLoading    bool
}

func NewResponsePanel(styles theme.Styles) ResponsePanel {
	vp := viewport.New(0, 0)
	vp.SetContent("No response yet. Press Ctrl+R to send request.")

	return ResponsePanel{
		styles:       styles,
		viewport:     vp,
		activeSubTab: RespTabBody,
		isPretty:     true,
	}
}

func (p *ResponsePanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.viewport.Width = w - 4
	p.viewport.Height = h - 6
}

func (p *ResponsePanel) SetFocused(focused bool) {
	p.focused = focused
}

func (p *ResponsePanel) SetLoading(loading bool) {
	p.isLoading = loading
	if loading {
		p.viewport.SetContent("Sending request... Please wait.")
	}
}

func (p *ResponsePanel) SetResponse(resp *models.Response) {
	p.Response = resp
	p.isLoading = false
	p.updateViewportContent()
}

func (p *ResponsePanel) updateViewportContent() {
	if p.Response == nil {
		p.viewport.SetContent("No response yet. Press Ctrl+R to send request.")
		return
	}

	if p.Response.Error != "" {
		p.viewport.SetContent(p.styles.StatusErr.Render(fmt.Sprintf("Error: %s", p.Response.Error)))
		return
	}

	switch p.activeSubTab {
	case RespTabBody:
		bodyStr := string(p.Response.Body)
		if len(bodyStr) == 0 {
			p.viewport.SetContent("(Empty Response Body)")
			return
		}

		if p.isPretty && strings.Contains(strings.ToLower(p.Response.ContentType), "json") {
			if pretty, err := theme.PrettyFormatJSON(p.Response.Body); err == nil {
				highlighted := theme.HighlightSyntax(pretty, "json", p.styles.Theme.ChromaStyleName)
				p.viewport.SetContent(highlighted)
				return
			}
		}

		p.viewport.SetContent(bodyStr)

	case RespTabHeaders:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Protocol: %s\n\n", p.Response.Proto))
		for _, h := range p.Response.Headers {
			sb.WriteString(p.styles.Key.Render(h.Key+": ") + p.styles.Value.Render(h.Value) + "\n")
		}
		p.viewport.SetContent(sb.String())

	case RespTabCookies:
		if len(p.Response.Cookies) == 0 {
			p.viewport.SetContent("No cookies returned in response.")
			return
		}
		var sb strings.Builder
		for _, c := range p.Response.Cookies {
			sb.WriteString(p.styles.Key.Render(c.Key+": ") + p.styles.Value.Render(c.Value) + "\n")
		}
		p.viewport.SetContent(sb.String())
	}
}

func (p ResponsePanel) Update(msg tea.Msg) (ResponsePanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !p.focused {
			return p, nil
		}

		switch msg.String() {
		case "alt+b":
			p.activeSubTab = RespTabBody
			p.updateViewportContent()
			return p, nil
		case "alt+h":
			p.activeSubTab = RespTabHeaders
			p.updateViewportContent()
			return p, nil
		case "alt+p":
			p.isPretty = !p.isPretty
			p.updateViewportContent()
			return p, nil
		}
	}

	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

func (p ResponsePanel) View() string {
	style := p.styles.Panel
	if p.focused {
		style = p.styles.ActivePanel
	}

	var statusHeader string
	if p.isLoading {
		statusHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Render(" Status: Requesting... ")
	} else if p.Response != nil {
		if p.Response.Error != "" {
			statusHeader = p.styles.StatusErr.Render(" Status: ERROR ")
		} else {
			statusStyle := p.styles.StatusStyle(p.Response.StatusCode)
			statusHeader = fmt.Sprintf(" Status: %s | Time: %d ms | Size: %d B ",
				statusStyle.Render(p.Response.StatusText),
				p.Response.ResponseTimeMs,
				p.Response.SizeBytes,
			)
		}
	} else {
		statusHeader = " Status: -- "
	}

	subTabs := []string{"Alt+B Body", "Alt+H Headers", "Cookies", "Alt+P Pretty"}
	var renderedTabs []string
	for i, t := range subTabs {
		if ResponseSubTab(i) == p.activeSubTab {
			renderedTabs = append(renderedTabs, p.styles.TabActive.Render(fmt.Sprintf("[%s]", t)))
		} else {
			renderedTabs = append(renderedTabs, p.styles.TabInactive.Render(t))
		}
	}
	tabBar := strings.Join(renderedTabs, "  ")

	topLine := lipgloss.JoinHorizontal(lipgloss.Center, statusHeader, " | ", tabBar)
	content := lipgloss.JoinVertical(lipgloss.Left, topLine, "\n", p.viewport.View())

	return style.
		Width(p.width - 2).
		Height(p.height - 2).
		Render(content)
}
