# apicli - Terminal API Client

A cross-platform terminal-based API client for testing and debugging HTTP APIs, built in Go using Bubble Tea, Lip Gloss, and Bubbles.

## Features

- **Fast & Lightweight**: Instant startup (< 300 ms) with minimal memory footprint.
- **Keyboard-First TUI**: Fully keyboard-driven 3-panel resizable layout.
- **HTTP Methods & Components**: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS with headers, query params, cookies, and body types (JSON, Raw, XML, Form URL Encoded, Multipart, Binary).
- **Environment & Variables**: Resolution of custom variables (`{{BASE_URL}}`, `{{TOKEN}}`) and built-in dynamic variables (`{{timestamp}}`, `{{uuid}}`, `{{randomInt}}`, `{{randomString}}`, `{{hostname}}`).
- **Response Viewer**: Pretty JSON formatting, raw response, syntax highlighting (via Chroma), status code badges, response timing metrics (ms), and size (bytes).
- **Import / Export**: Import cURL commands, OpenAPI/Swagger specifications, and Postman collections.
- **Request History**: Automatic request history logging with re-run capabilities.
- **Themes**: Dark, Light, and High Contrast themes.

## Keyboard Shortcuts

- `Ctrl+N` / `Ctrl+T`: New Request / New Tab
- `Ctrl+W`: Close Tab
- `Ctrl+R` / `Ctrl+Enter`: Send HTTP Request
- `Ctrl+S`: Save Request
- `Ctrl+F`: Search Requests
- `Ctrl+H`: View History
- `Ctrl+I`: Import cURL / OpenAPI
- `Ctrl+1` / `Ctrl+2` / `Ctrl+3`: Switch Focus between Collections, Request Editor, Response Viewer
- `Tab` / `Shift+Tab`: Cycle Panel Focus
- `Alt+1`..`4`: Switch Request Sub-Tabs (Params, Headers, Auth, Body)
- `Alt+B` / `Alt+H` / `Alt+P`: Switch Response Sub-Tabs / Toggle Pretty JSON
- `?`: Toggle Keyboard Reference Modal
- `Ctrl+Q`: Quit

## Installation & Usage

### Build Binary
```bash
go build -o apicli ./cmd/apicli
```

### Launch Interactive TUI
```bash
./apicli
```

### Direct CLI Execution
```bash
./apicli run "curl -X POST https://httpbin.org/post -d '{\"hello\":\"world\"}'"
```

### Import Collection
```bash
./apicli import path/to/openapi.yaml
```
