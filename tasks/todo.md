# Task List

- [ ] Task 1: Fix response panel bottom border calculation
  - Acceptance: Bottom border and rounded corners of the response panel are visible across all terminal heights without being cut off.
  - Verify: Run `go test ./...`, check visual rendering at different terminal heights.
  - Files: `internal/ui/app.go`, `internal/ui/response.go`

- [ ] Task 2: Implement line wrapping in Response Viewport and Request textareas
  - Acceptance: Characters exceeding panel width wrap onto subsequent lines instead of clipping; both long response bodies and request textareas wrap properly.
  - Verify: Load and send requests with long lines; verify all content wraps and scrolls.
  - Files: `internal/ui/response.go`, `internal/ui/editor.go`

- [ ] Checkpoint 1: Borders and Wrapping verified

- [ ] Task 3: Update Collection Model with BaseURL & Create Collection Modal
  - Acceptance: `models.Collection` includes `BaseURL string`; `ModalCreateCollection` prompts for Name and Base URL with Tab navigation and Enter confirmation.
  - Verify: Unit test collection storage with BaseURL; verify modal lifecycle in `modals.go`.
  - Files: `internal/models/models.go`, `internal/storage/storage.go`, `internal/ui/modals.go`

- [ ] Task 4: Build Sidebar Component (`internal/ui/sidebar.go`)
  - Acceptance: `SidebarPanel` displays collections and their requests; handles `up`/`down` navigation, selection highlights, `c` key to trigger collection creation, and `n` key for new request.
  - Verify: Unit test sidebar rendering and navigation state.
  - Files: `internal/ui/sidebar.go`

- [ ] Task 5: Integrate Sidebar into `AppModel` (`Alt+S` Toggle, Layout, Request Loading)
  - Acceptance: `Alt+S` toggles retractable sidebar; editor/response panel resize dynamically; requests created under collection inherit Base URL; pressing `Enter` on a request loads it into active tab.
  - Verify: `go test ./...`, run binary, verify keybindings, update Help modal and `README.md`.
  - Files: `internal/ui/app.go`, `internal/ui/modals.go`, `README.md`

- [ ] Checkpoint 2: Full End-to-End Verification
