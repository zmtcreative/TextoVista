<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

## v0.7.0 - 2026-10-19

This release significantly expands TextoVista's Markdown rendering capabilities with rich inline formatting extensions and comprehensive typographic shortcuts. Users can now leverage subscript and superscript notation, inline text highlighting, text insertion markup, and an extended set of automatic typographic replacements, including legal symbols, vulgar fractions, and directional arrows, all backed by dedicated configuration toggles in the Settings dialog and custom SCSS styling.

Additionally, Windows distribution and release automation have been streamlined with a new Scoop package manifest, consolidated NSIS installer packaging and hashing within the build scripts, and improved `package-lock.json` version synchronization. Project documentation has also been expanded with dedicated contribution guidelines, root-level architecture documentation, and an updated feature showcase.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### CHANGES

- Add: Added subscript (`~sub~`) and superscript (`^sup^`) Markdown support (`internal/markdown/markdown.go`, `github.com/bowman2001/goldmark-supersubscript`) with single-tilde discrimination to preserve GFM `~~strikethrough~~` parsing, enabled via the `UseSuperSubScript` configuration option and Settings toggle (`frontend/src/components/Settings.vue`).
- Add: Added inline text highlighting (`==highlight==` -> `<mark>`) support (`github.com/zuern/goldmark-highlight`) with a `UseInlineHighlighting` configuration option, Settings dialog toggle, and theme-aware SCSS styling (`frontend/src/assets/scss/_main.scss`).
- Add: Implemented a custom Goldmark text insertion extension (`++text++` -> `<ins>`) (`internal/gm-ext/insertion/insertion.go`) with a `UseTextInsertion` configuration option, Settings dialog toggle, and custom SCSS styling.
- Add: Implemented an internal typographic shortcuts extension (`internal/gm-ext/shortcuts/shortcuts.go`) supporting case-insensitive legal and technical symbols (`(c)`, `(r)`, `(tm)`, `(p)`, `(sm)`, `+-`), boundary-checked vulgar fractions (`1/2`, `1/4`, `3/4`), and ASCII arrows (`->`, `<-`, `<->`, `=>`, `<==`, `<=>`) with HTML comment and comparison exclusions.
- Add: Added a Scoop package manifest (`build/windows/scoop/TextoVista.json`) to support Windows portable installations.
- Add: Added comprehensive unit and feature matrix test suites covering subscript/superscript, inline highlighting, text insertion, and typographic shortcut parsing (`internal/markdown/supersub_test.go`, `internal/markdown/highlight_test.go`, `internal/markdown/insertion_test.go`, `internal/gm-ext/insertion/insertion_test.go`, `internal/gm-ext/shortcuts/shortcuts_test.go`).
- Update: Formatted the Go backend codebase with `gofmt` and updated Wails TypeScript models (`frontend/wailsjs/go/models.ts`) and test fixtures (`frontend/src/test-utils/config-fixtures.ts`) for the new Markdown configuration options.
- CI/CD: Streamlined the Windows packaging workflow (`.github/workflows/build.yml`) and Wails installer build script (`scripts/Build-WailsInstaller.ps1`) by consolidating NSIS installer compression and SHA-256 hashing directly into the build script.
- Scripts: Updated `scripts/Set-Version.ps1` to accurately synchronize version numbers in `frontend/package-lock.json`, and updated repository path references to `TextoVista-dev` in `scripts/Run-SyncToPublicRepo.ps1` and `scripts/verify-wsl-linux.sh`.
- Docs: Added contribution guidelines (`CONTRIBUTING.md`), promoted and updated system architecture documentation (`ARCHITECTURE.md`), expanded `README.md` with a comprehensive feature breakdown, and updated sample showcase documents (`assets/docs/sample.md`, `docs/sample1.md`, `docs/test-inline-formatting.md`) with examples of all new inline formatting and typographic extensions.

## v0.6.0 - 2026-10-18

This update delivers comprehensive security hardening, concurrency improvements, and enhanced document navigation across the Go backend and Vue frontend. Key architectural refinements strengthen IPC boundaries, prevent directory traversal and file resource exhaustion, and synchronize shared state across configuration, font, and theme managers to eliminate race conditions and resource leaks.

In addition, defense-in-depth HTML sanitization has been implemented using both backend policy rules and frontend `DOMPurify` integration while preserving rich HTML5, SVG diagram, and MathML formula rendering. User experience is further improved with a new navigation history stack and toolbar back button for traversing linked Markdown documents, resolved internal anchor scrolling behavior, and an expanded sample showcase document.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### CHANGES

- Add: Added document navigation history stack and a toolbar Back button (`frontend/src/components/Toolbar.vue`) to support navigating between linked Markdown files.
- Add: Added comprehensive backend (`app_test.go`, `internal/app/document_cover1_test.go`, `internal/gm-ext/htmlsanitize/htmlsanitize_test.go`) and frontend (`frontend/src/__tests__/App.spec.ts`) unit and regression test suites covering security boundaries, sanitization, and UI components.
- Update: Streamlined "Allow Inline HTML" configuration handling (`internal/app/config.go`, `frontend/src/components/Settings.vue`) and refined allowed link and sanitization policies.
- Update: Tracked `frontend/package-lock.json` for deterministic frontend dependency resolution.
- Fix: Hardened backend HTML sanitizer (`internal/gm-ext/htmlsanitize/htmlsanitize.go`) to block dangerous tags (`<object>`, `<embed>`, `<applet>`, `<base>`, `<meta>`), strip inline `on*` event attributes, and restrict URI schemes to `http`, `https`, `mailto`, `tel`, and `file`.
- Fix: Implemented frontend defense-in-depth sanitization with `DOMPurify` across rendered Markdown, document metadata, and frontmatter (`frontend/src/App.vue`, `frontend/src/components/FrontMatter.vue`) while preserving HTML5, SVG diagrams, and MathML formulas.
- Fix: Enforced IPC filesystem boundary validation in `LoadAndDisplayMarkdown` to restrict relative path traversal, reject binary files, and enforce a 20MB file size limit.
- Fix: Synchronized `ConfigManager`, `FontManager`, and `ThemeManager` with `sync.RWMutex` and bound file watcher debounce timers to cancellable contexts to prevent concurrent map panics and resource leaks.
- Fix: Resolved vertical scrolling and rendering issues when following internal document anchor links (`frontend/src/App.vue`, `frontend/src/assets/scss/_main.scss`).
- CI/CD: Updated the GitHub Actions build and release workflow (`.github/workflows/build.yml`) with concurrency group management to prevent conflicting workflow runs.
- Docs: Overhauled the embedded and repository sample Markdown showcase documents (`assets/docs/sample.md`, `docs/sample1.md`) with comprehensive feature examples, updated icon assets (`assets/docs/danger-square.svg`), and corrected image paths.

## v0.5.3 - 2027-10-18

This release marks the official rebranding of the application to **TextoVista** and delivers substantial updates across the runtime environment, dependencies, release automation, Linux packaging, and documentation.

Core language and framework requirements have been modernized to Go 1.27.0, Node.js >=22.0.0, and Wails v2.16.0. Markdown and diagram processing have been updated to use the official `d2lang/d2` rendering library alongside the latest Goldmark and syntax highlighting packages. In addition, release management has been overhauled with modular PowerShell 7 version-bumping (`scripts/Set-Version.ps1`) and automated changelog generation (`scripts/Update-Changelog.ps1`) scripts, while Linux build pipelines and GitHub Actions CI workflows have been refined for reliable cross-platform distribution.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### CHANGES

- Rebrand: Rebranded application name to **TextoVista** and updated the Go module (`texto-vista`), binary names, NSIS installer configuration, UI references, and icon assets.
- Add: Added automated changelog generation and update script (`scripts/Update-Changelog.ps1`).
- Scripts: Replaced `scripts/New-GitTag.ps1` with a modularized `scripts/Set-Version.ps1` featuring interactive tag collision resolution, clean-repository enforcement, and `-CurrentVersion` tagging support.
- Scripts: Updated WSL/Linux prerequisite validation (`scripts/verify-wsl-linux.sh`) and application build scripts (`scripts/build-wails-app.sh`) for improved Linux archive packaging.
- Update: Upgraded language and runtime requirements to Go 1.27.0 and Node.js >=22.0.0.
- Update: Upgraded desktop application framework to Wails v2.16.0 and updated HTML sanitization handling (`internal/gm-ext/htmlsanitize/htmlsanitize.go`).
- Update: Migrated D2 diagram rendering from `oss.terrastruct.com/d2` to `github.com/d2lang/d2` v0.9.0 and updated `goldmark-d2`.
- Update: Upgraded core Go dependencies (`goldmark` v1.8.6, `chroma/v2` v2.27.0, `gjson` v1.20.0, `regexp2/v2` v2.2.1, and `golang.org/x/*` packages).
- Update: Upgraded frontend dependencies, TypeScript, Vite tooling, and Playwright/Vitest test suites.
- Fix: Resolved Wails v2.16.0 version alignment and Linux packaging configuration across build scripts and GitHub Actions workflows.
- Docs: Updated system architecture (`docs/ARCHITECTURE.md`), WSL setup, and Linux packaging guides.

## v0.5.3 - 2026-10-18

This release marks the official rebranding of the application to **TextoVista** and delivers substantial updates across the runtime environment, dependencies, build automation, and documentation.

Core language and framework requirements have been modernized to Go 1.27, Node.js 22, and Wails v2.16.0. Markdown and diagram processing have been updated to use the official D2 rendering library alongside the latest Goldmark and syntax highlighting packages. In addition, Linux packaging workflows and automated version-management scripts have been overhauled for robust cross-platform development and CI/CD releases.

The application has been compiled and tested for Windows, and the tests and build completed successfully on both WSL2 (Ubuntu-26.04) and within the GitHub Actions CI workflow (generating standardized `.tar.gz` packages containing the Linux executable).

See the repository's [README](./README.md) for additional information

> [!WARNING]
>
> This is still **Beta** software! It should be functional for general usage, but development is still ongoing.

### CHANGES

- Rebrand: Rebranded application name to **TextoVista** and updated Go module, binary names, installer configuration, and application assets.
- Update: Upgraded language and runtime requirements to Go 1.27.0 and Node.js >=22.0.0.
- Update: Upgraded desktop application framework to Wails v2.16.0 and adjusted HTML sanitization handling.
- Update: Migrated D2 diagram rendering to `github.com/d2lang/d2` v0.9.0 and updated `goldmark-d2`.
- Update: Upgraded core Go modules (`goldmark` v1.8.6, `chroma/v2` v2.27.0, `gjson` v1.20.0, and `golang.org/x/*` packages).
- Update: Upgraded frontend packages, TypeScript, Vite tooling, and test suites (Playwright and Vitest).
- CI/CD: Enhanced GitHub Actions workflows and Linux packaging scripts (`scripts/build-wails-app.sh`) for Go 1.27 and consistent `.tar.gz` distribution artifact naming.
- Docs: Updated system architecture (`docs/ARCHITECTURE.md`), Linux packaging guides, and project documentation.

## v0.5.2 - 2026-07-05

The application provides configuration options to enable/disable various Markdown syntax enhancements.

This release includes full configuration settings options and includes the ability for the user to select the proportional and monospace font of their choice and set the default font size for each. It has also had some refinements to the internal CSS to improve rendering, though more work needs to be done. Very basic printing support is available as well, though the CSS formatting for printing still needs tweaking.

The application has been compiled and tested for Windows, and the tests and build completed successfully on both WSL2 (Ubuntu-26.04) and within the GitHub CI workflow (to generate a `.tar.gz` package containing the Linux executable). The Go and Wails frameworks are multi-platform and this application could (in theory) be compiled on MacOS as well, but this hasn't been tested at this time.

See the repository's [README](./README.md) for additional information

> [!WARNING]
>
> This is still **Beta** software! It should be functional for general usage, but development is still ongoing.

### CHANGES

- Add: Manual/Auto document refresh (Auto-refresh enabled by default) -- refreshes currently loaded Markdown file when a change is detected to the file.
- Fix: Various security and maintenance fixes.

## v0.5.1 - 2026-06-05

This is the first public release of this Markdown Reader application. It supports most CommonMark specs and GitHub Markdown syntax.

The application provides configuration options to enable/disable various Markdown syntax enhancements.

This release includes full configuration settings options and includes the ability for the user to select the proportional and monospace font of their choice and set the default font size for each. It has also had some refinements to the internal CSS to improve rendering, though more work needs to be done. Very basic printing support is available as well, though the CSS formatting for printing still needs tweaking.

The application has been compiled and tested for Windows, and the tests and build completed successfully on both WSL2 (Ubuntu-26.04) and within the GitHub CI workflow (to generate a `.tar.gz` package containing the Linux executable). The Go and Wails frameworks are multi-platform and this application could (in theory) be compiled on MacOS as well, but this hasn't been tested at this time.

See the repository's [README](./README.md) for additional information

> [!WARNING]
>
> This is still **Beta** software! It should be functional for general usage, but development is still ongoing.
