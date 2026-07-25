package ui

import (
	"fmt"

	"vapeu/internal/models"
	"vapeu/internal/theme"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type CollectionItemDelegate struct {
	Styles theme.Styles
}

type item struct {
	title, desc string
	node        *models.CollectionNode
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title + " " + i.desc }

type CollectionsPanel struct {
	collections []models.Collection
	list        list.Model
	styles      theme.Styles
	width       int
	height      int
	focused     bool
	flatNodes   []*models.CollectionNode
}

func NewCollectionsPanel(collections []models.Collection, styles theme.Styles) CollectionsPanel {
	var items []list.Item
	var flatNodes []*models.CollectionNode

	for _, col := range collections {
		for _, node := range col.Nodes {
			flatNodes = append(flatNodes, node)
			desc := "Folder"
			if node.Kind == models.NodeRequest && node.Request != nil {
				desc = fmt.Sprintf("[%s] %s", node.Request.Method, node.Request.URL)
			}
			items = append(items, item{
				title: node.Name,
				desc:  desc,
				node:  node,
			})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 10, 10)
	l.Title = "Collections"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	return CollectionsPanel{
		collections: collections,
		list:        l,
		styles:      styles,
		flatNodes:   flatNodes,
	}
}

func (p *CollectionsPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.list.SetSize(w-2, h-2)
}

func (p *CollectionsPanel) SetFocused(focused bool) {
	p.focused = focused
}

func (p CollectionsPanel) Update(msg tea.Msg) (CollectionsPanel, tea.Cmd) {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p *CollectionsPanel) SelectedRequest() *models.Request {
	selected := p.list.SelectedItem()
	if selected == nil {
		return nil
	}
	itm, ok := selected.(item)
	if !ok || itm.node == nil {
		return nil
	}
	return itm.node.Request
}

func (p CollectionsPanel) View() string {
	style := p.styles.Panel
	if p.focused {
		style = p.styles.ActivePanel
	}

	return style.
		Width(p.width - 2).
		Height(p.height - 2).
		Render(p.list.View())
}

func (p *CollectionsPanel) Refresh(collections []models.Collection) {
	p.collections = collections
	var items []list.Item
	var flatNodes []*models.CollectionNode

	for _, col := range collections {
		for _, node := range col.Nodes {
			flatNodes = append(flatNodes, node)
			desc := "Folder"
			if node.Kind == models.NodeRequest && node.Request != nil {
				desc = fmt.Sprintf("[%s] %s", node.Request.Method, node.Request.URL)
			}
			items = append(items, item{
				title: fmt.Sprintf("📁 %s / %s", col.Name, node.Name),
				desc:  desc,
				node:  node,
			})
		}
	}

	p.flatNodes = flatNodes
	p.list.SetItems(items)
}
