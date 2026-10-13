package ui

import (
	"fmt"
	"strings"

	"vapeu/internal/models"
	"vapeu/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SidebarItemType int

const (
	SidebarItemCollection SidebarItemType = iota
	SidebarItemRequest
)

type SidebarItem struct {
	Type           SidebarItemType
	CollectionIdx  int
	NodeIdx        int
	CollectionID   string
	CollectionName string
	BaseURL        string
	Request        *models.Request
}

type SidebarPanel struct {
	styles        theme.Styles
	width         int
	height        int
	focused       bool
	collections   []models.Collection
	items         []SidebarItem
	selectedIdx   int
	collapsedCols map[string]bool
}

func NewSidebarPanel(styles theme.Styles, collections []models.Collection) SidebarPanel {
	sb := SidebarPanel{
		styles:        styles,
		collections:   collections,
		width:         28,
		collapsedCols: make(map[string]bool),
	}
	sb.rebuildItems()
	return sb
}

func (s *SidebarPanel) SetCollections(cols []models.Collection) {
	s.collections = cols
	s.rebuildItems()
}

func (s *SidebarPanel) SetSize(w, h int) {
	s.width = w
	s.height = h
}

func (s *SidebarPanel) SetFocused(focused bool) {
	s.focused = focused
}

func (s *SidebarPanel) rebuildItems() {
	var items []SidebarItem
	for cIdx, col := range s.collections {
		items = append(items, SidebarItem{
			Type:           SidebarItemCollection,
			CollectionIdx:  cIdx,
			CollectionID:   col.ID,
			CollectionName: col.Name,
			BaseURL:        col.BaseURL,
		})

		if s.collapsedCols != nil && s.collapsedCols[col.ID] {
			continue
		}

		for nIdx, node := range col.Nodes {
			if node.Request != nil {
				items = append(items, SidebarItem{
					Type:          SidebarItemRequest,
					CollectionIdx: cIdx,
					NodeIdx:       nIdx,
					CollectionID:  col.ID,
					BaseURL:       col.BaseURL,
					Request:       node.Request,
				})
			}
		}
	}
	s.items = items
	if s.selectedIdx >= len(s.items) {
		s.selectedIdx = len(s.items) - 1
	}
	if s.selectedIdx < 0 && len(s.items) > 0 {
		s.selectedIdx = 0
	}
}

func (s *SidebarPanel) SelectedItem() *SidebarItem {
	if s.selectedIdx >= 0 && s.selectedIdx < len(s.items) {
		return &s.items[s.selectedIdx]
	}
	return nil
}

func (s SidebarPanel) Update(msg tea.Msg) (SidebarPanel, tea.Cmd, *models.Request, bool, bool, *SidebarItem) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || !s.focused {
		return s, nil, nil, false, false, nil
	}

	var (
		cmd           tea.Cmd
		reqToLoad     *models.Request
		triggerNewCol bool
		triggerNewReq bool
		reqToDelete   *SidebarItem
	)

	switch keyMsg.String() {
	case "up", "k":
		if s.selectedIdx > 0 {
			s.selectedIdx--
		}

	case "down", "j":
		if s.selectedIdx < len(s.items)-1 {
			s.selectedIdx++
		}

	case "enter", " ":
		item := s.SelectedItem()
		if item != nil {
			if item.Type == SidebarItemCollection {
				if s.collapsedCols == nil {
					s.collapsedCols = make(map[string]bool)
				}
				s.collapsedCols[item.CollectionID] = !s.collapsedCols[item.CollectionID]
				s.rebuildItems()
			} else if item.Type == SidebarItemRequest {
				reqToLoad = item.Request
			}
		}

	case "c":
		triggerNewCol = true

	case "n":
		triggerNewReq = true

	case "d", "x", "delete":
		item := s.SelectedItem()
		if item != nil && item.Type == SidebarItemRequest {
			reqToDelete = item
		}
	}

	return s, cmd, reqToLoad, triggerNewCol, triggerNewReq, reqToDelete
}

func (s SidebarPanel) View() string {
	style := s.styles.Panel
	if s.focused {
		style = s.styles.ActivePanel
	}

	titleText := " 📂 Collections "
	header := lipgloss.NewStyle().Bold(true).Foreground(s.styles.Theme.TabActive).Render(titleText)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render(" [c]New [n]+Req [d]Del [␣]Fold")

	topLine := lipgloss.JoinHorizontal(lipgloss.Left, header, hint)

	var sb strings.Builder
	sb.WriteString(topLine + "\n")

	if len(s.items) == 0 {
		sb.WriteString("\n " + lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086")).Render("No collections yet.\n Press 'c' to create."))
	} else {
		maxVisible := s.height - 4
		if maxVisible < 1 {
			maxVisible = 1
		}

		start := 0
		if s.selectedIdx >= maxVisible {
			start = s.selectedIdx - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(s.items) {
			end = len(s.items)
		}

		for i := start; i < end; i++ {
			item := s.items[i]
			cursor := "  "
			if i == s.selectedIdx {
				cursor = "> "
			}

			if item.Type == SidebarItemCollection {
				colTitle := item.CollectionName
				if colTitle == "" {
					colTitle = "Unnamed"
				}
				// Truncate if too long for sidebar width
				maxColLen := s.width - 10
				if maxColLen > 3 && len(colTitle) > maxColLen {
					colTitle = colTitle[:maxColLen-2] + ".."
				}

				icon := "▼"
				if s.collapsedCols != nil && s.collapsedCols[item.CollectionID] {
					icon = "▶"
				}

				line := fmt.Sprintf("%s%s 📁 %s", cursor, icon, colTitle)
				if i == s.selectedIdx {
					sb.WriteString(s.styles.TabActive.Bold(true).Render(line) + "\n")
				} else {
					sb.WriteString(lipgloss.NewStyle().Foreground(s.styles.Theme.Foreground).Render(line) + "\n")
				}
			} else {
				reqName := item.Request.Name
				if reqName == "" {
					reqName = item.Request.URL
				}
				maxReqLen := s.width - 14
				if maxReqLen > 3 && len(reqName) > maxReqLen {
					reqName = reqName[:maxReqLen-2] + ".."
				}

				mBadge := s.styles.MethodStyle(item.Request.Method).Render(item.Request.Method[:1])
				line := fmt.Sprintf("%s  %s %s", cursor, mBadge, reqName)
				if i == s.selectedIdx {
					sb.WriteString(lipgloss.NewStyle().Foreground(s.styles.Theme.ActiveBorder).Bold(true).Render(line) + "\n")
				} else {
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")).Render(line) + "\n")
				}
			}
		}
	}

	content := sb.String()

	return style.
		Width(s.width - 2).
		Height(s.height - 2).
		Render(content)
}
