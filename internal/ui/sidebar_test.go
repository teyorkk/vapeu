package ui

import (
	"testing"

	"vapeu/internal/models"
	"vapeu/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestSidebarFolding(t *testing.T) {
	styles := theme.MakeStyles(theme.DarkTheme)
	cols := []models.Collection{
		{
			ID:   "col-1",
			Name: "Collection 1",
			Nodes: []*models.CollectionNode{
				{
					ID:      "node-1",
					Name:    "Req 1",
					Kind:    models.NodeRequest,
					Request: &models.Request{ID: "req-1", Name: "Req 1", Method: "GET", URL: "http://example.com/1"},
				},
				{
					ID:      "node-2",
					Name:    "Req 2",
					Kind:    models.NodeRequest,
					Request: &models.Request{ID: "req-2", Name: "Req 2", Method: "POST", URL: "http://example.com/2"},
				},
			},
		},
	}

	sb := NewSidebarPanel(styles, cols)
	sb.SetFocused(true)

	// Initially expanded: 1 collection item + 2 request items = 3 items
	assert.Len(t, sb.items, 3)
	assert.Equal(t, SidebarItemCollection, sb.items[0].Type)
	assert.Equal(t, SidebarItemRequest, sb.items[1].Type)
	assert.Equal(t, SidebarItemRequest, sb.items[2].Type)

	// Selection is on collection (index 0). Press Space to fold/collapse.
	sb, _, reqToLoad, _, _, _ := sb.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	assert.Nil(t, reqToLoad)
	// Now collapsed: only 1 collection item
	assert.Len(t, sb.items, 1)
	assert.True(t, sb.collapsedCols["col-1"])

	viewCollapsed := sb.View()
	assert.Contains(t, viewCollapsed, "▶")

	// Press Enter to expand again
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Len(t, sb.items, 3)
	assert.False(t, sb.collapsedCols["col-1"])

	viewExpanded := sb.View()
	assert.Contains(t, viewExpanded, "▼")
}

func TestSidebarDeleteRequest(t *testing.T) {
	styles := theme.MakeStyles(theme.DarkTheme)
	cols := []models.Collection{
		{
			ID:   "col-1",
			Name: "Collection 1",
			Nodes: []*models.CollectionNode{
				{
					ID:      "node-1",
					Name:    "Req 1",
					Kind:    models.NodeRequest,
					Request: &models.Request{ID: "req-1", Name: "Req 1", Method: "GET", URL: "http://example.com/1"},
				},
			},
		},
	}

	sb := NewSidebarPanel(styles, cols)
	sb.SetFocused(true)

	// Move cursor down to the request
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, sb.selectedIdx)
	assert.Equal(t, SidebarItemRequest, sb.SelectedItem().Type)

	// Press 'd' to trigger delete
	_, _, _, _, _, reqToDelete := sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.NotNil(t, reqToDelete)
	assert.Equal(t, "req-1", reqToDelete.Request.ID)
}

func TestSidebarMultipleCollectionsFolding(t *testing.T) {
	styles := theme.MakeStyles(theme.DarkTheme)
	cols := []models.Collection{
		{
			ID:   "col-1",
			Name: "Collection 1",
			Nodes: []*models.CollectionNode{
				{
					ID:      "node-1",
					Name:    "Req 1",
					Kind:    models.NodeRequest,
					Request: &models.Request{ID: "req-1", Name: "Req 1", Method: "GET", URL: "http://example.com/1"},
				},
				{
					ID:      "node-2",
					Name:    "Req 2",
					Kind:    models.NodeRequest,
					Request: &models.Request{ID: "req-2", Name: "Req 2", Method: "POST", URL: "http://example.com/2"},
				},
			},
		},
		{
			ID:   "col-2",
			Name: "Collection 2",
			Nodes: []*models.CollectionNode{
				{
					ID:      "node-3",
					Name:    "Req 3",
					Kind:    models.NodeRequest,
					Request: &models.Request{ID: "req-3", Name: "Req 3", Method: "GET", URL: "http://example.com/3"},
				},
			},
		},
	}

	sb := NewSidebarPanel(styles, cols)
	sb.SetFocused(true)

	// Initially: Col 1 (1) + 2 reqs (2) + Col 2 (1) + 1 req (1) = 5 items
	assert.Len(t, sb.items, 5)

	// Fold Col 1
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	assert.True(t, sb.collapsedCols["col-1"])
	assert.False(t, sb.collapsedCols["col-2"])
	assert.Len(t, sb.items, 3) // Col 1 (collapsed) + Col 2 + Req 3

	// Move cursor down to Col 2 (index 1)
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, sb.selectedIdx)
	assert.Equal(t, "col-2", sb.items[sb.selectedIdx].CollectionID)

	// Fold Col 2
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	assert.True(t, sb.collapsedCols["col-2"])
	assert.Len(t, sb.items, 2) // Col 1 + Col 2 only

	// Unfold Col 1: move up to Col 1 (index 0) and press Enter
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, sb.selectedIdx)
	sb, _, _, _, _, _ = sb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, sb.collapsedCols["col-1"])
	assert.True(t, sb.collapsedCols["col-2"])
	assert.Len(t, sb.items, 4) // Col 1 + 2 reqs + Col 2 (collapsed)
}
