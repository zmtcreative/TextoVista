# TextoVista Architecture

## Overview

This is a desktop Markdown reader application built with **Wails v2** (Go backend + Vue.js frontend). It provides extensive markdown rendering capabilities with numerous extensions and features.

## Technology Stack

### Backend (Go)
- **Framework**: Wails v2.16.0
- **Go Version**: 1.27.0
- **Markdown Parser**: Goldmark (yuin/goldmark v1.8.6)
- **Configuration**: Viper (spf13/viper v1.21.0)
- **File Watching**: fsnotify (fsnotify/fsnotify v1.10.1)
- **Key Libraries**:
  - goldmark-highlighting (syntax highlighting with Chroma v2.27.0)
  - goldmark-katex (math rendering v2.1.0)
  - goldmark-mermaid (diagram rendering v0.6.0)
  - goldmark-d2 (D2 diagrams)
  - Custom extensions: gm-alert-callouts (v0.8.0), gm-fancy-lists (v0.3.0), gm-sectionwrapper (v0.3.0)
  - mdigger/goldmark-attributes & nemunaire/goldmark-inline-attributes

### Frontend (Vue.js)
- **Framework**: Vue 3.5.44 with TypeScript 6.0.3
- **Runtime / Engine**: Node.js >= 22.0.0 (CI uses Node 24)
- **Build Tool**: Vite 8.3.4
- **Testing**: Vitest 5.0.5 (unit), Playwright 1.64.1 (e2e)
- **Styling**: SCSS (sass 1.105.2) with PostCSS (postcss 8.5.30, postcss-nesting 14.0.3, autoprefixer 10.6.1)
- **Key Libraries**:
  - Mermaid.js 12.2.0 (diagrams)
  - KaTeX 0.19.0 (math rendering)
  - DOMPurify 3.4.17 (HTML sanitization)

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Desktop Application                      │
│                      (Wails Runtime Bridge)                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼────────┐                    ┌────────▼────────┐
│  Go Backend     │                    │  Vue Frontend   │
│                 │◄─── Bindings ─────►│                 │
└───────┬─────────┘                    └────────┬────────┘
        │                                       │
        │ Events                                │ Events
        └───────────────────────────────────────┘
```

## Backend Architecture (Go)

### Main Entry Point
```
main.go
├── CLI Argument Parsing (internal/cli)
├── App Initialization (app.go)
└── Wails App Startup
```

### Core Application Structure (app.go)
```
App Struct
├── Context (Wails context)
├── Current File Path
├── Frontmatter Data
├── Managers:
│   ├── ThemeManager (theme.go)
│   ├── PrintManager (print.go)
│   ├── FileManager (file.go)
│   ├── DocumentProcessor (document.go)
│   ├── BinaryDetector (binary.go)
│   ├── ConfigManager (config.go)
│   └── FontManager (font_manager.go)
├── File Watcher (fsnotify)
└── Auto-refresh Timer
```

### Configuration Management (internal/app/config.go)
```
ConfigManager (Viper)
├── Application Options:
│   ├── Inline HTML Support
│   ├── HTML Sanitization
│   ├── Strip H1
│   ├── Frontmatter Title
│   ├── Auto-refresh
│   └── Font Settings (family, size, mono)
├── Markdown Options:
│   ├── GFM Support
│   ├── PHP Markdown Extensions
│   ├── Emoji Support
│   ├── Mermaid Diagrams
│   ├── KaTeX Math
│   ├── D2 Diagrams
│   ├── Syntax Highlighting
│   ├── Fancy Lists
│   ├── Attributes
│   ├── Typographic Extensions
│   ├── Inline Text Highlighting
│   ├── Text Insertion
│   └── Abbreviations
└── Alert Callout Options:
    ├── Enable Callouts
    └── Style (GFMStrict, GFMWithAliases, GFMPlus, Obsidian)
```

### Document Processing Pipeline (internal/app/document.go)
```
DocumentProcessor
│
├── 1. File Reading
│   ├── Binary Detection
│   ├── UTF-16 BOM Detection & Conversion
│   └── Line Ending Normalization
│
├── 2. Markdown Conversion (internal/markdown/markdown.go)
│   ├── Goldmark Instance Creation
│   ├── Extension Loading (based on config)
│   ├── Frontmatter Parsing
│   ├── Abbreviation Extraction
│   ├── H1 Title Extraction
│   └── HTML Generation
│
├── 3. Metadata Processing
│   ├── Frontmatter Title/Date/Type
│   ├── File Modification Date
│   └── Title Priority (frontmatter > H1 > filename)
│
├── 4. HTML Cleanup
│   ├── Whitespace Normalization
│   ├── H1 Stripping (if configured)
│   └── Abbreviation Replacement
│
├── 5. Frontmatter HTML Formatting
│   └── Structured HTML generation
│
└── 6. Event Emission
    ├── markdown-rendered event
    ├── add/remove/toggle-doc-class events
    └── error events
```

### Markdown Processing (internal/markdown/markdown.go)
```
Goldmark Extensions Pipeline
│
├── Core Parser
│   └── Auto Heading IDs
│
├── Optional Extensions (config-driven):
│   ├── GFM (GitHub Flavored Markdown)
│   ├── PHP Markdown Extensions (Definition Lists, Footnotes)
│   ├── Typographer (fancy quotes, dashes)
│   ├── Frontmatter (YAML/TOML)
│   ├── Mermaid (diagrams)
│   ├── Emoji
│   ├── Figure (image captions)
│   ├── Anchor (heading links)
│   ├── Fences (fenced DIVs)
│   ├── Section Wrapper (nested sections)
│   ├── Syntax Highlighting (Chroma)
│   ├── Fancy Lists (Pandoc-style)
│   ├── Alert Callouts (custom icons)
│   ├── HTML Sanitization (custom)
│   ├── Attributes (block & inline)
│   ├── KaTeX (math)
│   └── D2 Diagrams
│
└── Custom Renderer
    ├── HTML Output
    └── Code Block Attributes
```

### File Management (internal/app/file.go)
```
FileManager
├── File Dialog Integration (Wails runtime)
├── File Loading
├── Binary File Detection
└── Error Handling
```

### Font Management (internal/app/font_manager.go)
```
FontManager
├── System Font Detection
│   ├── Basic Detection (Windows/Linux)
│   └── Advanced Detection (registry scanning)
├── Monospace Font Detection
├── Font Validation
└── Font List Caching
```

### Theme Management (internal/app/theme.go)
```
ThemeManager
├── Theme Storage (light/dark)
├── Theme Getter/Setter
└── Theme Change Events
```

### Print Management (internal/app/print.go)
```
PrintManager
├── Print Content (system print dialog)
└── Print to PDF (Windows-specific)
```

### File Watching (app.go)
```
File Watcher (fsnotify)
├── Directory Watching
├── File Change Detection
├── Debounced Auto-refresh (250ms)
└── Path Normalization (Windows case-insensitive)
```

## Frontend Architecture (Vue.js)

### Main Application (App.vue)
```
App.vue
├── Template:
│   ├── Toolbar Component
│   ├── FrontMatter Component
│   ├── Header (Title, Date, Error)
│   ├── Main Content (rendered HTML)
│   ├── Help Dialog
│   └── Settings Dialog
│
├── Script Setup:
│   ├── State Management (refs)
│   ├── Wails Event Listeners:
│   │   ├── markdown-rendered
│   │   ├── error
│   │   ├── theme:changed
│   │   ├── show-help
│   │   ├── show-settings
│   │   ├── add/remove/toggle-doc-class
│   │   └── print-content
│   ├── Wails Method Calls:
│   │   ├── GetTheme/SetTheme
│   │   ├── GetCurrentFont/GetCurrentMonospaceFont
│   │   ├── HasCurrentFile
│   │   └── ReloadCurrentDocument
│   ├── Mermaid Initialization
│   ├── KaTeX Rendering
│   └── Font Settings Application
│
└── Styling:
    ├── App Container Layout
    ├── Error Message Styling
    └── Print Overrides
```

### Components

#### Toolbar.vue
```
Toolbar
├── Theme Toggle (light/dark)
├── Refresh Button
├── Frontmatter Toggle
└── Current Theme Display
```

#### FrontMatter.vue
```
FrontMatter
├── Frontmatter HTML Display
├── Visibility Toggle
└── Collapsible Container
```

#### Settings.vue
```
Settings Dialog
├── Tabs:
│   ├── Application Settings
│   │   ├── Inline HTML
│   │   ├── Sanitize HTML
│   │   ├── Strip H1
│   │   ├── Frontmatter Title
│   │   ├── Auto-refresh
│   │   └── Font Settings (family, size, mono)
│   ├── Markdown Features
│   │   ├── GFM
│   │   ├── PHP Extensions
│   │   ├── Emoji
│   │   ├── Mermaid
│   │   ├── KaTeX
│   │   ├── D2 Diagrams
│   │   ├── Syntax Highlighting
│   │   ├── Fancy Lists
│   │   ├── Attributes
│   │   ├── Typographic
│   │   ├── Inline Text Highlighting
│   │   ├── Text Insertion
│   │   └── Abbreviations
│   └── Alert Callouts
│       ├── Enable
│       └── Style Selection
├── Save Options:
│   ├── Save to Config File
│   └── Session Only
└── Font Pickers
    ├── Proportional Font
    └── Monospace Font
```

#### Help.vue
```
Help Dialog
├── Title Display
├── Content Display (HTML)
├── Close Button
└── Modal Behavior
```

#### Icon.vue
```
Icon Component
├── SVG Icon Rendering
└── Dynamic Icon Loading
```

### Communication Flow

#### Backend to Frontend (Events)
```
Go Backend                    Vue Frontend
     │                              │
     │ EventsEmit()                 │ EventsOn()
     │                              │
     ├─► markdown-rendered ────────►│ Update HTML
     ├─► error ───────────────────►│ Show Error
     ├─► theme:changed ───────────►│ Update Theme
     ├─► show-help ────────────────►│ Open Help Dialog
     ├─► show-settings ────────────►│ Open Settings Dialog
     ├─► add-doc-class ───────────►│ Add CSS Class
     ├─► remove-doc-class ────────►│ Remove CSS Class
     ├─► toggle-doc-class ───────►│ Toggle CSS Class
     └─► print-content ───────────►│ Trigger Print
```

#### Frontend to Backend (Method Calls)
```
Vue Frontend                  Go Backend
     │                              │
     │ Wails Binding                │ Exported Method
     │                              │
     ├─► GetTheme() ──────────────►│ Return Theme
     ├─► SetTheme() ──────────────►│ Set Theme
     ├─► GetSettings() ──────────►│ Return Config
     ├─► SaveSettings() ──────────►│ Save Config
     ├─► SaveSettingsSessionOnly()─►│ Update Session Config
     ├─► LoadAndDisplayMarkdown()─►│ Process & Render
     ├─► ReloadCurrentDocument()─►│ Reload Current File
     ├─► GetCurrentFont() ────────►│ Return Font Settings
     ├─► SetApplicationFont() ───►│ Set Font
     ├─► GetAvailableFonts() ─────►│ Return Font List
     ├─► GetMonospaceFontsWithInfo()►│ Return Mono Fonts
     ├─► GetAlertCalloutStyles() ─►│ Return Styles
     └─► PrintContent() ───────────►│ Trigger Print
```

## Data Flow

### File Loading Flow
```
User Action (File > Open or CLI)
    │
    ├─► FileManager.OpenFileMenuHandler()
    │       │
    │       ├─► Wails File Dialog
    │       │
    │       └─► DocumentProcessor.LoadAndDisplayMarkdown()
    │               │
    │               ├─► Read File (with encoding detection)
    │               │
    │               ├─► markdown.ConvertMarkdownToHTML()
    │               │       │
    │               │       ├─► Parse Frontmatter
    │               │       ├─► Extract H1
    │               │       ├─► Apply Extensions
    │               │       └─► Generate HTML
    │               │
    │               ├─► Process Metadata
    │               │
    │               ├─► Format Frontmatter HTML
    │               │
    │               ├─► Cleanup HTML
    │               │
    │               └─► EventsEmit("markdown-rendered")
    │                       │
    │                       └─► Vue Frontend Receives
    │                               │
    │                               ├─► Update renderedHTML
    │                               ├─► Update Title/Date
    │                               ├─► Render Mermaid Diagrams
    │                               ├─► Render KaTeX Math
    │                               └─► Apply Font Settings
```

### Settings Change Flow
```
User Action (Settings Dialog)
    │
    ├─► Settings.vue
    │       │
    │       ├─► GetSettings() (load current)
    │       │
    │       ├─► User Modifies Settings
    │       │
    │       └─► SaveSettings() or SaveSettingsSessionOnly()
    │               │
    │               ├─► ConfigManager.SetConfig()
    │               │
    │               ├─► ConfigManager.SaveConfig() (if persistent)
    │               │
    │               ├─► Recreate DocumentProcessor
    │               │       │
    │               │       └─► New Goldmark Instance with new config
    │               │
    │               ├─► Recreate FileManager
    │               │
    │               └─► ReloadCurrentDocument()
    │                       │
    │                       └─► Re-render with new settings
```

### Auto-refresh Flow
```
File Change on Disk
    │
    ├─► fsnotify Event
    │       │
    │       ├─► handleWatchedFileEvent()
    │       │       │
    │       │       ├─► Validate File Path
    │       │       │
    │       │       └─► scheduleAutoRefresh()
    │       │               │
    │       │               ├─► Debounce (250ms)
    │       │               │
    │       │               └─► reloadCurrentDocument()
    │       │                       │
    │       │                       └─► DocumentProcessor.LoadAndDisplayMarkdown()
    │       │                               │
    │       │                               └─► EventsEmit("markdown-rendered")
```

## Key Features Implementation

### Markdown Extensions
- **GFM**: Tables, strikethrough, autolinks, task lists
- **PHP Markdown**: Definition lists, footnotes
- **Typographic**: Fancy quotes, en/em-dashes, ellipsis
- **Diagrams**: Mermaid.js, D2 diagrams
- **Math**: KaTeX rendering
- **Code**: Syntax highlighting with Chroma
- **Alert Callouts**: GFM and Obsidian styles with custom icons
- **Attributes**: Block and inline attributes
- **Fancy Lists**: Pandoc-style (letters, roman numerals)
- **Sections**: Nested section wrapping
- **Inline Styling**: Text highlighting (`==text==`), text insertion (`++text++`), superscript (`^text^`), and subscript (`~text~`)

### Font Management
- **Detection**: Basic and advanced monospace font detection
- **Customization**: Separate proportional and monospace font settings
- **Persistence**: Saved to config file
- **Dynamic**: CSS custom properties for real-time updates

### Theme Support
- **Light/Dark Themes**: Toggleable via toolbar and backend
- **Mermaid Theme**: Syncs with app theme
- **CSS Variables**: Theme-aware styling

### File Watching
- **Real-time**: Auto-refresh on file changes
- **Debounced**: 250ms delay to prevent excessive reloads
- **Configurable**: Can be disabled in settings

## Testing Architecture

TextoVista maintains a comprehensive, layered testing strategy combining Go backend tests, Vue frontend unit tests, and Playwright end-to-end integration tests with Wails dev server lifecycle orchestration.

### 1. Go Backend Tests
- **Location**: Throughout `internal/` packages (`internal/app/`, `internal/cli/`, `internal/markdown/`, `internal/utils/`, `internal/gm-ext/htmlsanitize/`).
- **Scope**: Configuration loading/saving, CLI argument parsing, document processing pipeline, frontmatter extraction, custom Goldmark extensions, HTML sanitization, and utility functions.
- **Execution**:
  ```bash
  go test ./...
  ```

### 2. Frontend Unit Tests (Vitest)
- **Framework**: Vitest with `@vue/test-utils` and `jsdom`.
- **Location**: `frontend/src/__tests__/` and `frontend/src/components/__tests__/`.
- **Scope**: Vue component rendering, event emission, settings tab behaviors, frontmatter visibility, and DOM updates.
- **Execution**:
  ```bash
  cd frontend
  npm run test:unit
  ```

### 3. End-to-End Tests (Playwright & WailsDevHelper)
- **Framework**: Playwright with Chromium.
- **Helper Architecture (`frontend/tests/wails-dev-helper.ts`)**:
  - Automatically launches the Wails development server (`wails dev`) from the project root.
  - Buffers process `stdout` and `stderr` streams, extracting the local development URL (e.g., `http://localhost:5173/` or internal port) with detailed failure reporting if the server fails to start.
  - Launches and attaches Playwright Chromium in either `headless` or `interactive` mode.
  - Shares a single running Wails application instance across test cases within a suite for high execution speed.
  - Provides fast in-memory and `page.reload()` state resetting between tests without requiring costly app restarts.
  - Handles graceful application teardown via Wails `window.runtime.Quit()` followed by process signal termination (`SIGINT` / `SIGKILL`).
- **Test Suites (`frontend/tests/`)**:
  - `main-test-suite.spec.ts`: Core application display, responsiveness, Help modal interactions, theme toggling, UI consistency, and performance checkpoints (`headless-safe`).
  - `fast-sequential-tests.spec.ts`: Rapid state-reset verification demonstrating performance benefits of shared application instances (`headless-safe`).
  - `interactive-shortcuts.spec.ts`: Native keyboard shortcuts and OS-level menu interactions (`interactive-native`).
- **Runtime Modes**:
  - Controlled via environment variables: `TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE` (or legacy `MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE`).
  - Defaults to `headless` in CI (`CI=true`) and `interactive` in local desktop development when unspecified.

### 4. Test Automation Scripts
- **PowerShell (Windows)**:
  - `scripts/Run-FrontendTests.ps1`: Orchestrates Playwright suites with options:
    - `-TestSuite [main|fast|interactive|all]`
    - `-RuntimeMode [auto|headless|interactive]`
    - `-ShowReport`
  - `scripts/Run-AllTests.ps1`: Executes the full test matrix (Go tests + Playwright suites).
- **Bash (Linux / WSL)**:
  - `scripts/run-all-tests.sh`: Runs Go and frontend tests in Linux environments.
- **Test Impact Caching**:
  - `scripts/Build-WailsInstaller.ps1` maintains Git commit-based test outcome caching in `tmp/test-cache/build-wailsinstaller-tests.json`. Builds skip redundant test runs if the repository state matches the cached tested commit and no source files have changed.

## CI/CD Pipeline & Release Automation

The project uses GitHub Actions ([.github/workflows/build.yml](file:///c:/Users/dougl/Projects/TextoVista-dev/.github/workflows/build.yml)) for continuous integration, multi-platform artifact generation, and automated releases.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           GitHub Actions Workflow                               │
└───────────────────────┬─────────────────────────────────┬───────────────────────┘
                        │                                 │
           ┌────────────▼────────────┐       ┌────────────▼────────────┐
           │   test-windows (pwsh)   │       │    test-linux (bash)    │
           │  • Go 1.27 + Node 24    │       │  • Go 1.27 + Node 24    │
           │  • Wails CLI v2.16.0    │       │  • Wails CLI v2.16.0    │
           │  • Go Unit Tests        │       │  • GTK3 & WebKit2GTK    │
           │  • Playwright Headless  │       │  • Linux Wails Build    │
           └────────────┬────────────┘       └────────────┬────────────┘
                        │ (on tags / dispatch)            │ (on tags / dispatch)
           ┌────────────▼────────────┐       ┌────────────▼────────────┐
           │      build-windows      │       │   build-linux-tar-gz    │
           │  • NSIS Installer (.exe)│       │  • Linux Binary (.tar.gz│
           │  • ZIP + Checksums      │       │  • SHA256 & SHA1 Hashes │
           └────────────┬────────────┘       └────────────┬────────────┘
                        │                                 │
                        └───────────────┬─────────────────┘
                                        │
                           ┌────────────▼────────────┐
                           │      release (v*)       │
                           │  • CHANGELOG.md Notes   │
                           │  • GitHub Release Assets│
                           └─────────────────────────┘
```

### Workflow Jobs
1. **`test-windows`**:
   - Runs on `windows-latest`.
   - Provisions Go 1.27 (via `go-version-file: 'go.mod'`) and Node.js 24.
   - Installs Wails CLI `v2.16.0` (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`).
   - Runs Go unit tests (`go test ./...`).
   - Installs Playwright Chromium and executes `./scripts/Run-FrontendTests.ps1 -TestSuite main`.
2. **`test-linux`**:
   - Runs on `ubuntu-latest`.
   - Installs Linux GUI dependencies (`build-essential`, `pkg-config`, `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`).
   - Installs Wails CLI `v2.16.0`.
   - Validates Linux compilation via `wails build -tags webkit2_41`.
3. **`build-windows`**:
   - Builds the standalone Windows application and NSIS installer via `scripts/Build-WailsInstaller.ps1 -NSIS -NoTests`.
   - Packages installer into a ZIP archive and computes SHA256 and SHA1 checksums.
   - Uploads build artifacts (`texto-vista-windows`).
4. **`build-linux-tar-gz`**:
   - Compiles Linux binary with `webkit2_41` tags.
   - Stages and packages the release archive `textovista-<version>-linux-amd64.tar.gz` with SHA256/SHA1 checksums.
   - Uploads build artifacts (`textovista-linux`).
5. **`release`**:
   - Triggered when pushing tags matching `v*` or via manual `workflow_dispatch`.
   - Downloads Windows and Linux artifacts.
   - Parses the version section from `CHANGELOG.md` corresponding to the tag.
   - Publishes the GitHub Release via `softprops/action-gh-release@v3` with all installers, archives, and checksums attached.

## Build Process & Version Management

### Development Workflow
```bash
# Run backend and frontend together with hot reload
wails dev

# Or launch using helper scripts
./scripts/Run-WailsDev.ps1       # Windows
./scripts/run-wails-dev.sh       # Linux / WSL
```

### Production Build Scripts
- **Windows**: `scripts/Build-WailsInstaller.ps1`
  - Supports flags: `-Build`, `-NSIS`, `-UPX` (executable compression), `-ZipFile`, `-RunAllTests`, `-NoTests`.
  - Automatically extracts Git commit hash, date stamp, and release tags to inject ldflags:
    `-ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit} -s -w"`.
- **Linux**: `scripts/build-wails-app.sh`
  - Compiles the application targeting WebKit2GTK 4.1 (`wails build -tags webkit2_41`).

### Version Synchronization (`scripts/Set-Version.ps1`)
The repository uses a centralized, modular PowerShell 7 script to maintain version consistency across all project metadata files:
- `wails.json` (`info.productVersion`)
- `frontend/package.json` (`version`)
- `frontend/package-lock.json` (`version` and root package version)
- `build/windows/info.json` (`ProductVersion`, `FileVersion`, `LegalCopyright`)
- `build/windows/installer/project.nsi` (`PRODUCT_VERSION`, `FILE_VERSION`)

Supports semantic version increments (`-Major`, `-Minor`, `-Patch`), setting explicit versions, or tagging the existing version (`-CurrentVersion`) with automated Git tag creation (`Version: <tag>` annotation).

## Configuration Management

Configuration is handled through **Viper** and stored in JSON format:
- **File Name**: `<appName>.json` (e.g. `texto-vista.json`, or `texto-vista-dev.json` when running in development mode)
- **Directory**:
  - **Windows**: `%APPDATA%\<appName>\` (e.g. `C:\Users\<user>\AppData\Roaming\texto-vista\`)
  - **Linux**: `$XDG_CONFIG_HOME/<appName>/` or `~/.config/<appName>/`
- **Structure**:
  - `application`: Inline HTML support, HTML sanitization, strip H1, frontmatter title priority, auto-refresh toggle, font family/size, monospace font family/size, and advanced font detection.
  - `markdown`: Toggles for GFM, PHP Markdown Extra, emoji, mermaid, figure, anchor headings, fences, section wrapping, Chroma syntax highlighting, fancy lists, attributes, typographic punctuation, KaTeX math, D2 diagrams, inline text highlighting, text insertion, superscript/subscript, and abbreviations.
  - `alert_callouts`: Enable/disable alert callouts and style selection (`GFMStrict`, `GFMWithAliases`, `GFMPlus`, `Obsidian`).

## Summary

TextoVista combines a robust Go backend and a modern Vue 3 frontend connected through Wails v2:
- **High Performance**: Native OS windowing with minimal footprint and instant rendering.
- **Extensible Parser**: Goldmark modular extension pipeline supporting diagrams, math, callouts, and styling attributes.
- **Resilient CI/CD & Testing**: Headless Playwright integration with automatic dev server management, cross-platform builds (Windows NSIS + Linux tar.gz), and checksum-verified releases.
- **Enterprise-Ready Configuration**: User-configurable markdown features, font customization, and seamless live reloading via file system notifications.
