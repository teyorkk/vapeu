package ui

import (
	"os"
	"path/filepath"
	"testing"

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
