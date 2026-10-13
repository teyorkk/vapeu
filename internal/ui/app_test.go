package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vapeu/internal/models"
	"vapeu/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestAppModelInitialization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	app := NewAppModel(st)
	app.showSplash = false
	assert.NotNil(t, app.storage)
	assert.Len(t, app.tabs, 1)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	app.width = msg.Width
	app.height = msg.Height
	app.recalculateLayout()

	viewOutput := app.View()
	assert.NotEmpty(t, viewOutput)
	lines := strings.Split(viewOutput, "\n")
	assert.LessOrEqual(t, len(lines), app.height-2, "total view lines must be at most app.height - 2 to preserve safe bottom border margin")
	assert.Contains(t, lines[0], "VAPEU", "line 0 must contain VAPEU branding header")
	assert.Contains(t, lines[0], "Help", "line 0 must contain Help hint")
	t.Logf("Line 0: %q", lines[0])
	t.Logf("Total lines: %d", len(lines))
}

func TestSidebarToggleAndCollectionBaseURL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-sidebar-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	app := NewAppModel(st)
	app.showSplash = false
	app.width = 120
	app.height = 40
	app.recalculateLayout()

	// Initially sidebar is open
	assert.True(t, app.showSidebar)

	// Press Alt+S to retract sidebar -> focus must switch directly to FocusEditor (request panel)
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	app = model.(AppModel)
	assert.False(t, app.showSidebar)
	assert.Equal(t, FocusEditor, app.activeFocus, "closing sidebar must focus request panel")

	// Press Alt+S to reopen sidebar -> focus must switch to FocusSidebar (collections)
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	app = model.(AppModel)
	assert.True(t, app.showSidebar)
	assert.Equal(t, FocusSidebar, app.activeFocus, "opening sidebar must focus collections sidebar")

	// Create collection with BaseURL
	newCol := &models.Collection{
		ID:      "test-col-1",
		Name:    "Test Collection",
		BaseURL: "https://api.example.com/v1",
	}
	app.collections = append(app.collections, *newCol)
	app.sidebarPanel.SetCollections(app.collections)
	assert.Len(t, app.collections, 2) // demo + new

	// Select the newly added collection in sidebar
	app.sidebarPanel.selectedIdx = len(app.sidebarPanel.items) - 1
	app.createRequestUnderCurrentCollection()

	// Verify the newly created request has URL prepended with BaseURL
	currentReq := app.editorPanel.Request
	assert.NotNil(t, currentReq)
	assert.True(t, len(currentReq.URL) >= len(newCol.BaseURL))
	assert.Contains(t, currentReq.URL, "https://api.example.com/v1/")
}

func TestDeleteRequestInApp(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-delete-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	app := NewAppModel(st)
	app.showSplash = false
	app.width = 120
	app.height = 40
	app.recalculateLayout()

	// Initial collection has 1 node
	assert.NotEmpty(t, app.collections)
	initialNodeCount := len(app.collections[0].Nodes)
	assert.True(t, initialNodeCount > 0)

	// Set focus to sidebar and select the request (index 1)
	app.activeFocus = FocusSidebar
	app.updateFocus()
	app.sidebarPanel.selectedIdx = 1 // request item
	assert.Equal(t, SidebarItemRequest, app.sidebarPanel.SelectedItem().Type)
	targetReqID := app.sidebarPanel.SelectedItem().Request.ID

	// Press 'd' to initiate delete request -> must show ModalConfirmDelete
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	assert.Equal(t, ModalConfirmDelete, app.modals.ActiveModal, "pressing 'd' must open confirmation modal")

	// Press 'y' to confirm deletion
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	app = model.(AppModel)
	assert.Equal(t, ModalNone, app.modals.ActiveModal, "confirming must close modal")

	// Verify request was removed from collection
	assert.Len(t, app.collections[0].Nodes, initialNodeCount-1)

	// Verify storage was updated
	loadedCols, err := st.LoadCollections()
	assert.NoError(t, err)
	assert.Len(t, loadedCols[0].Nodes, initialNodeCount-1)

	// Verify target request ID is gone
	for _, node := range loadedCols[0].Nodes {
		if node.Request != nil {
			assert.NotEqual(t, targetReqID, node.Request.ID)
		}
	}
}

func TestTerminalHeightMatrix(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-height-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	testHeights := []int{15, 20, 24, 30, 40, 50, 60}
	for _, h := range testHeights {
		app := NewAppModel(st)
		app.showSplash = false
		app.width = 120
		app.height = h
		app.recalculateLayout()

		view := app.View()
		assert.NotEmpty(t, view)
		lines := strings.Split(view, "\n")
		assert.True(t, len(lines) <= h-2, "rendered line count must never exceed h - 2 for terminal height %d (got %d)", h, len(lines))
		assert.Contains(t, lines[0], "VAPEU", "line 0 must contain branding at height %d", h)
	}
}

func TestDeleteRequestActiveTabSync(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-tabsync-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	app := NewAppModel(st)
	app.showSplash = false
	app.width = 120
	app.height = 40
	app.recalculateLayout()

	req1 := &models.Request{ID: "req-1", Name: "Req 1", Method: "GET", URL: "http://example.com/1"}
	req2 := &models.Request{ID: "req-2", Name: "Req 2", Method: "POST", URL: "http://example.com/2"}

	col := models.Collection{
		ID:   "col-1",
		Name: "Col 1",
		Nodes: []*models.CollectionNode{
			{ID: "node-1", Name: "Req 1", Kind: models.NodeRequest, Request: req1},
			{ID: "node-2", Name: "Req 2", Kind: models.NodeRequest, Request: req2},
		},
	}
	app.collections = []models.Collection{col}
	app.sidebarPanel.SetCollections(app.collections)

	// Set up 2 tabs: Tab 0 (Req 1), Tab 1 (Req 2)
	app.tabs = []tabItem{
		{Title: req1.Name, Request: req1},
		{Title: req2.Name, Request: req2},
	}
	app.activeTab = 0
	app.editorPanel.LoadRequest(req1)

	// In sidebar, select req-1 (index 1) and delete it
	app.activeFocus = FocusSidebar
	app.updateFocus()
	app.sidebarPanel.selectedIdx = 1
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	assert.Equal(t, ModalConfirmDelete, app.modals.ActiveModal)

	// Confirm with 'y'
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	app = model.(AppModel)

	// Verify Tab 0 (Req 1) was removed, now only Tab (Req 2) remains
	assert.Len(t, app.tabs, 1)
	assert.Equal(t, "Req 2", app.tabs[0].Title)
	assert.Equal(t, req2.ID, app.editorPanel.Request.ID)
}

func TestDeleteRequestCancelFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-cancel-delete-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	app := NewAppModel(st)
	app.showSplash = false
	app.width = 120
	app.height = 40
	app.recalculateLayout()

	initialNodeCount := len(app.collections[0].Nodes)

	// Focus sidebar and press 'd'
	app.activeFocus = FocusSidebar
	app.updateFocus()
	app.sidebarPanel.selectedIdx = 1

	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	assert.Equal(t, ModalConfirmDelete, app.modals.ActiveModal)

	// Press 'n' to cancel
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	app = model.(AppModel)
	assert.Equal(t, ModalNone, app.modals.ActiveModal)
	assert.Len(t, app.collections[0].Nodes, initialNodeCount, "request must not be deleted when cancelled with 'n'")

	// Press 'd' again and press 'esc' to cancel
	app.activeFocus = FocusSidebar
	app.updateFocus()
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app = model.(AppModel)
	assert.Equal(t, ModalConfirmDelete, app.modals.ActiveModal)

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = model.(AppModel)
	assert.Equal(t, ModalNone, app.modals.ActiveModal)
	assert.Len(t, app.collections[0].Nodes, initialNodeCount, "request must not be deleted when cancelled with 'esc'")
}

func TestTelescopeDeleteRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "apicli-ui-telescope-del-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	st, err := storage.NewStorage(filepath.Join(tempDir, ".apiclient"))
	assert.NoError(t, err)

	app := NewAppModel(st)
	app.showSplash = false
	app.width = 120
	app.height = 40
	app.recalculateLayout()

	req1 := &models.Request{ID: "req-t1", Name: "Telescope Req 1", Method: "GET", URL: "http://example.com/t1"}
	req2 := &models.Request{ID: "req-t2", Name: "Telescope Req 2", Method: "POST", URL: "http://example.com/t2"}

	col := models.Collection{
		ID:   "col-t",
		Name: "Col Telescope",
		Nodes: []*models.CollectionNode{
			{ID: "node-t1", Name: "Telescope Req 1", Kind: models.NodeRequest, Request: req1},
			{ID: "node-t2", Name: "Telescope Req 2", Kind: models.NodeRequest, Request: req2},
		},
	}
	_ = st.SaveCollection(col)
	app.collections = []models.Collection{col}
	app.sidebarPanel.SetCollections(app.collections)
	app.tabs = []tabItem{{Title: req1.Name, Request: req1}, {Title: req2.Name, Request: req2}}
	app.activeTab = 0
	app.editorPanel.LoadRequest(req1)

	// Open Telescope
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeySpace, Alt: true})
	app = model.(AppModel)
	assert.Equal(t, ModalSearch, app.modals.ActiveModal)
	assert.Len(t, app.modals.allRequests, 2)

	// Press Ctrl+D on selected request (req-t1)
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	app = model.(AppModel)
	assert.Equal(t, ModalConfirmDelete, app.modals.ActiveModal, "Ctrl+D in Telescope must open ModalConfirmDelete")
	assert.Equal(t, "req-t1", app.modals.pendingDeleteReq.ID)

	// Confirm with 'y'
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	app = model.(AppModel)
	assert.Equal(t, ModalNone, app.modals.ActiveModal, "confirming must close modal")

	// Verify request was removed from collection
	assert.Len(t, app.collections[0].Nodes, 1)
	assert.Equal(t, "req-t2", app.collections[0].Nodes[0].Request.ID)

	// Verify storage persistence
	loadedCols, err := st.LoadCollections()
	assert.NoError(t, err)
	var foundCol *models.Collection
	for i := range loadedCols {
		if loadedCols[i].ID == col.ID {
			foundCol = &loadedCols[i]
			break
		}
	}
	assert.NotNil(t, foundCol)
	assert.Len(t, foundCol.Nodes, 1)
	assert.Equal(t, "req-t2", foundCol.Nodes[0].Request.ID)

	// Verify tab sync
	assert.Len(t, app.tabs, 1)
	assert.Equal(t, "Telescope Req 2", app.tabs[0].Title)

	// Verify Telescope cache updated
	assert.Len(t, app.modals.allRequests, 1)
	assert.Equal(t, "req-t2", app.modals.allRequests[0].ID)
}


