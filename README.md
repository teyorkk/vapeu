# VAPEU - Terminal API Client

> A fast, keyboard-first, cross-platform terminal API client built in Go using Bubble Tea, Lip Gloss, and Bubbles.

---

## ✨ Key Features

- **🚀 Fast & Lightweight**: Startup in under 300 ms with minimal CPU/RAM footprint. Works locally and over SSH.
- **⚡ Neovim Telescope Finder (`Alt+Space`)**: Instant fuzzy search across all your saved and historical requests by name, URL, or method.
- **🖥️ Clean 2-Panel UI**: Full-width vertical split featuring a Request Editor top panel and Response Viewer bottom panel.
- **🎨 Live Color Themes (`Alt+K`)**: Cycle through built-in themes instantly: **Dark**, **Light**, **Cyberpunk**, **Nord**, and **High Contrast**.
- **🔑 Interactive Auth (`Alt+3`)**: Live configuration for Bearer Tokens, Basic Auth (User/Password), and API Keys (`Alt+A`).
- **🛑 Stop / Cancel Requests (`Ctrl+X`)**: Abort long-running requests on demand with context cancellation.
- **🌐 HTTP/2 & Smart Redirects**: High-performance HTTP/2 transport with connection pooling and automatic 307/308 redirect following.
- **📦 Multi-Format Request Body**: JSON, Raw Text, XML, Form URL Encoded, Multipart Form Data, and Binary files.
- **🔍 Response Viewer**: Pretty JSON formatting, raw response, syntax highlighting (via Chroma), status code badges, response time (ms), and size metrics.
- **📥 cURL / OpenAPI / Postman Import (`Alt+I`)**: Import cURL commands, OpenAPI/Swagger specifications, and Postman collections directly.
- **📜 Scoped & Dynamic Variables**: Support for `{{BASE_URL}}`, `{{TOKEN}}`, and dynamic functions (`{{uuid}}`, `{{timestamp}}`, `{{randomInt}}`, `{{randomString}}`, `{{hostname}}`).

---

## ⌨️ Keyboard Shortcuts

### General & Navigation
| Shortcut | Action |
|---|---|
| `Alt+Space` / `Ctrl+F` | 🔭 Open Telescope Request Finder |
| `Alt+Left` / `Alt+Right` (`Alt+[` / `Alt+]`) | Switch Open Request Tabs |
| `Ctrl+N` / `Ctrl+T` | Create New Request / New Tab |
| `Ctrl+W` | Close Current Tab |
| `Tab` / `Shift+Tab` | Toggle Focus (Request Editor $\leftrightarrow$ Response Viewer) |
| `Ctrl+1` / `Ctrl+2` | Direct Focus (1: Editor, 2: Response) |
| `Alt+K` | 🎨 Cycle UI Color Theme |
| `?` | Show Keyboard Reference Modal |
| `Ctrl+Q` | Quit Application |

### Request Operations
| Shortcut | Action |
|---|---|
| `Ctrl+R` / `Ctrl+Enter` | Send HTTP Request |
| `Ctrl+X` | 🛑 Stop / Cancel Running Request |
| `Ctrl+S` | Save Request (Prompts for Name) |
| `Alt+I` / `Ctrl+I` | Import cURL / OpenAPI Specification |
| `Ctrl+H` | View Request History |

### Sub-Tab Navigation
| Shortcut | Action |
|---|---|
| `Alt+1` | Request Params (Query `key=value`) |
| `Alt+2` | Request Headers (`Header: Value`) |
| `Alt+3` | Request Auth (`Alt+A` to toggle None / Bearer / Basic / APIKey) |
| `Alt+4` | Request Body (JSON / Raw / Form) |
| `Alt+5` | Request Cookies (`cookie_name=value`) |
| `ESC` / `Alt+U` | Focus URL Bar (from sub-tab text areas) |
| `Alt+B` | Response Body Sub-Tab |
| `Alt+H` | Response Headers Sub-Tab |
| `Alt+C` | Response Cookies Sub-Tab |
| `Alt+P` | Toggle Pretty JSON Formatting |

---

## 🛠️ Installation & Usage

### Build Binary
```bash
go build -o apicli ./cmd/apicli
```

### Launch Interactive TUI
```bash
./apicli
```

### Run Direct CLI Request
```bash
./apicli run "curl -X POST https://jsonplaceholder.typicode.com/posts -d '{\"title\":\"foo\",\"body\":\"bar\",\"userId\":1}'"
```

### Import OpenAPI / Postman Specs
```bash
./apicli import path/to/openapi.yaml
```

---

## ⚙️ Configuration & Storage

VAPEU stores configuration and collection data locally in your home directory:
- Config: `~/.apicli/config.yaml`
- Collections: `~/.apicli/collections/`
- Environments: `~/.apicli/environments/`
- History: `~/.apicli/history.json`
