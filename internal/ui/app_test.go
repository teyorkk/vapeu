package ui

import (
	"os"
	"path/filepath"
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
	assert.NotNil(t, app.storage)
	assert.Len(t, app.tabs, 1)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	app.width = msg.Width
	app.height = msg.Height
	app.recalculateLayout()

	viewOutput := app.View()
	assert.NotEmpty(t, viewOutput)
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

	// Press Alt+S to retract sidebar
	model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	app = model.(AppModel)
	assert.False(t, app.showSidebar)

	// Press Alt+S to reopen sidebar
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	app = model.(AppModel)
	assert.True(t, app.showSidebar)

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

