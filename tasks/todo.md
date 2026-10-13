# Task List

- [x] Task 1: Correct Sizing in EditorPanel and ResponsePanel
  - Acceptance: Outer line counts of EditorPanel and ResponsePanel strictly match their assigned heights.
  - Verify: Unit test verifies `len(strings.Split(editorView, "\n")) == editorPanel.height` and `len(strings.Split(responseView, "\n")) == responsePanel.height`.
  - Files: `internal/ui/editor.go`, `internal/ui/response.go`

- [x] Task 2: Correct Sizing in SidebarPanel and AppModel.recalculateLayout
  - Acceptance: SidebarPanel renders exactly `availHeight` lines; total rendered lines of `AppModel.View()` equals `m.height`. Top header bar is at line 0.
  - Verify: Unit test verifies `len(strings.Split(app.View(), "\n")) == app.height`.
  - Files: `internal/ui/sidebar.go`, `internal/ui/app.go`, `internal/ui/app_test.go`

- [x] Checkpoint: Verify top header and bottom borders are completely visible simultaneously
