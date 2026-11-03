<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

This release introduces application window state persistence, enabling TextoVista to remember and restore window dimensions and maximized state across sessions while clamping bounds to the active display work area with Windows DPI scaling support. Additionally, the application's visual identity has been refreshed with updated vector, PNG, and ICO icon assets.

Release automation and developer tooling have also been refined with Keep a Changelog 1.1.0 standardization across changelog and version management scripts, SSH pre-flight verification and tag-based synchronization defaults in the public repository sync script, collapsible change logs in CI release notes, path exclusions for IDE and agent configuration directories in build workflows, and preconfigured MinGW64 CGO settings for local Windows development.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### Added

- Persistent window dimension and state management ([`app.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/app.go), [`main.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/main.go), [`internal/app/config.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/internal/app/config.go), [`frontend/wailsjs/go/models.ts`](file:///C:/Users/dougl/Projects/TextoVista-dev/frontend/wailsjs/go/models.ts)) to save and restore window width, height, and maximized state across application launches.
- Display work area detection with Windows DPI scaling support ([`internal/app/screen.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/internal/app/screen.go), [`internal/app/screen_windows.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/internal/app/screen_windows.go), [`internal/app/screen_other.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/internal/app/screen_other.go)) to clamp restored window dimensions and prevent off-screen positioning on startup, backed by unit test suites ([`internal/app/screen_test.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/internal/app/screen_test.go), [`internal/app/config_test.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/internal/app/config_test.go), [`main_test.go`](file:///C:/Users/dougl/Projects/TextoVista-dev/main_test.go)).
- SSH pre-flight verification functions ([`Assert-SshAgentRunning`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Run-SyncToPublicRepo.ps1) and [`Assert-GitRemoteAccess`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Run-SyncToPublicRepo.ps1)) in [`scripts/Run-SyncToPublicRepo.ps1`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Run-SyncToPublicRepo.ps1) to validate agent status, loaded keys, and remote connectivity prior to cloning.
- GitHub Sponsors funding configuration ([`.github/FUNDING.yml`](file:///C:/Users/dougl/Projects/TextoVista-dev/.github/FUNDING.yml)).

### Changed

- Standardized [`CHANGELOG.md`](file:///C:/Users/dougl/Projects/TextoVista-dev/CHANGELOG.md) formatting, category subheadings, and version compare links to adhere to Keep a Changelog 1.1.0, and updated [`scripts/Update-Changelog.ps1`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Update-Changelog.ps1), [`scripts/Set-Version.ps1`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Set-Version.ps1), and [`.github/workflows/build.yml`](file:///C:/Users/dougl/Projects/TextoVista-dev/.github/workflows/build.yml) to parse bracketed version headers and maintain reference link definitions.
- Defaulted public repository synchronization ([`scripts/Run-SyncToPublicRepo.ps1`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Run-SyncToPublicRepo.ps1)) to target the latest SemVer release tag via [`Get-LatestReleaseTag`](file:///C:/Users/dougl/Projects/TextoVista-dev/scripts/Run-SyncToPublicRepo.ps1#L176-L184) and synchronize tags by default, with `-IncludeUnreleasedMain` and `-SkipTagSync` switch parameters for override control.
- Refreshed the TextoVista SVG logo and regenerated PNG and Windows ICO icon assets ([`assets/icon/svg/texto-vista-icon.svg`](file:///C:/Users/dougl/Projects/TextoVista-dev/assets/icon/svg/texto-vista-icon.svg), [`assets/icon/texto-vista-icon.png`](file:///C:/Users/dougl/Projects/TextoVista-dev/assets/icon/texto-vista-icon.png), [`assets/icon/texto-vista.ico`](file:///C:/Users/dougl/Projects/TextoVista-dev/assets/icon/texto-vista.ico)), archiving legacy Markdown Reader icons to `assets/icon/old/` and updating [`assets/icon/make-icon-mac.cmd`](file:///C:/Users/dougl/Projects/TextoVista-dev/assets/icon/make-icon-mac.cmd).
- Configured the GitHub Actions build workflow ([`.github/workflows/build.yml`](file:///C:/Users/dougl/Projects/TextoVista-dev/.github/workflows/build.yml)) to ignore changes in `.vscode/**`, `.agent/**`, and `.gemini/**` on push events, and wrapped the `### CHANGES` section of generated release notes in a collapsible `<details>` block.
- Configured VS Code workspace terminal settings ([`.vscode/settings.json`](file:///C:/Users/dougl/Projects/TextoVista-dev/.vscode/settings.json)) to prepend MinGW64 to `PATH` and enable `CGO_ENABLED=1` by default for local Windows builds.

## [0.7.0] - 2026-10-19

This release significantly expands TextoVista's Markdown rendering capabilities with rich inline formatting extensions and comprehensive typographic shortcuts. Users can now leverage subscript and superscript notation, inline text highlighting, text insertion markup, and an extended set of automatic typographic replacements, including legal symbols, vulgar fractions, and directional arrows, all backed by dedicated configuration toggles in the Settings dialog and custom SCSS styling.

Additionally, Windows distribution and release automation have been streamlined with a new Scoop package manifest, consolidated NSIS installer packaging and hashing within the build scripts, and improved `package-lock.json` version synchronization. Project documentation has also been expanded with dedicated contribution guidelines, root-level architecture documentation, and an updated feature showcase.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### Added

- Subscript (`~sub~`) and superscript (`^sup^`) Markdown support (`internal/markdown/markdown.go`, `github.com/bowman2001/goldmark-supersubscript`) with single-tilde discrimination to preserve GFM `~~strikethrough~~` parsing, enabled via the `UseSuperSubScript` configuration option and Settings toggle (`frontend/src/components/Settings.vue`).
- Inline text highlighting (`==highlight==` -> `<mark>`) support (`github.com/zuern/goldmark-highlight`) with a `UseInlineHighlighting` configuration option, Settings dialog toggle, and theme-aware SCSS styling (`frontend/src/assets/scss/_main.scss`).
- Custom Goldmark text insertion extension (`++text++` -> `<ins>`) (`internal/gm-ext/insertion/insertion.go`) with a `UseTextInsertion` configuration option, Settings dialog toggle, and custom SCSS styling.
- Internal typographic shortcuts extension (`internal/gm-ext/shortcuts/shortcuts.go`) supporting case-insensitive legal and technical symbols (`(c)`, `(r)`, `(tm)`, `(p)`, `(sm)`, `+-`), boundary-checked vulgar fractions (`1/2`, `1/4`, `3/4`), and ASCII arrows (`->`, `<-`, `<->`, `=>`, `<==`, `<=>`) with HTML comment and comparison exclusions.
- Scoop package manifest (`build/windows/scoop/TextoVista.json`) to support Windows portable installations.
- Comprehensive unit and feature matrix test suites covering subscript/superscript, inline highlighting, text insertion, and typographic shortcut parsing (`internal/markdown/supersub_test.go`, `internal/markdown/highlight_test.go`, `internal/markdown/insertion_test.go`, `internal/gm-ext/insertion/insertion_test.go`, `internal/gm-ext/shortcuts/shortcuts_test.go`).
- Contribution guidelines (`CONTRIBUTING.md`), promoted and updated system architecture documentation (`ARCHITECTURE.md`), expanded `README.md` with a comprehensive feature breakdown, and updated sample showcase documents (`assets/docs/sample.md`, `docs/sample1.md`, `docs/test-inline-formatting.md`) with examples of all new inline formatting and typographic extensions.

### Changed

- Formatted the Go backend codebase with `gofmt` and updated Wails TypeScript models (`frontend/wailsjs/go/models.ts`) and test fixtures (`frontend/src/test-utils/config-fixtures.ts`) for the new Markdown configuration options.
- Streamlined the Windows packaging workflow (`.github/workflows/build.yml`) and Wails installer build script (`scripts/Build-WailsInstaller.ps1`) by consolidating NSIS installer compression and SHA-256 hashing directly into the build script.
- Updated `scripts/Set-Version.ps1` to accurately synchronize version numbers in `frontend/package-lock.json`, and updated repository path references to `TextoVista-dev` in `scripts/Run-SyncToPublicRepo.ps1` and `scripts/verify-wsl-linux.sh`.

## [0.6.0] - 2026-10-18

This update delivers comprehensive security hardening, concurrency improvements, and enhanced document navigation across the Go backend and Vue frontend. Key architectural refinements strengthen IPC boundaries, prevent directory traversal and file resource exhaustion, and synchronize shared state across configuration, font, and theme managers to eliminate race conditions and resource leaks.

In addition, defense-in-depth HTML sanitization has been implemented using both backend policy rules and frontend `DOMPurify` integration while preserving rich HTML5, SVG diagram, and MathML formula rendering. User experience is further improved with a new navigation history stack and toolbar back button for traversing linked Markdown documents, resolved internal anchor scrolling behavior, and an expanded sample showcase document.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### Added

- Document navigation history stack and a toolbar Back button (`frontend/src/components/Toolbar.vue`) to support navigating between linked Markdown files.
- Comprehensive backend (`app_test.go`, `internal/app/document_cover1_test.go`, `internal/gm-ext/htmlsanitize/htmlsanitize_test.go`) and frontend (`frontend/src/__tests__/App.spec.ts`) unit and regression test suites covering security boundaries, sanitization, and UI components.

### Changed

- Streamlined "Allow Inline HTML" configuration handling (`internal/app/config.go`, `frontend/src/components/Settings.vue`) and refined allowed link and sanitization policies.
- Tracked `frontend/package-lock.json` for deterministic frontend dependency resolution.
- Updated the GitHub Actions build and release workflow (`.github/workflows/build.yml`) with concurrency group management to prevent conflicting workflow runs.
- Overhauled the embedded and repository sample Markdown showcase documents (`assets/docs/sample.md`, `docs/sample1.md`) with comprehensive feature examples, updated icon assets (`assets/docs/danger-square.svg`), and corrected image paths.

### Fixed

- Hardened backend HTML sanitizer (`internal/gm-ext/htmlsanitize/htmlsanitize.go`) to block dangerous tags (`<object>`, `<embed>`, `<applet>`, `<base>`, `<meta>`), strip inline `on*` event attributes, and restrict URI schemes to `http`, `https`, `mailto`, `tel`, and `file`.
- Implemented frontend defense-in-depth sanitization with `DOMPurify` across rendered Markdown, document metadata, and frontmatter (`frontend/src/App.vue`, `frontend/src/components/FrontMatter.vue`) while preserving HTML5, SVG diagrams, and MathML formulas.
- Enforced IPC filesystem boundary validation in `LoadAndDisplayMarkdown` to restrict relative path traversal, reject binary files, and enforce a 20MB file size limit.
- Synchronized `ConfigManager`, `FontManager`, and `ThemeManager` with `sync.RWMutex` and bound file watcher debounce timers to cancellable contexts to prevent concurrent map panics and resource leaks.
- Resolved vertical scrolling and rendering issues when following internal document anchor links (`frontend/src/App.vue`, `frontend/src/assets/scss/_main.scss`).

## [0.5.3] - 2026-10-18

This release marks the official rebranding of the application to **TextoVista** and delivers substantial updates across the runtime environment, dependencies, release automation, Linux packaging, and documentation.

Core language and framework requirements have been modernized to Go 1.27.0, Node.js >=22.0.0, and Wails v2.16.0. Markdown and diagram processing have been updated to use the official `d2lang/d2` rendering library alongside the latest Goldmark and syntax highlighting packages. In addition, release management has been overhauled with modular PowerShell 7 version-bumping (`scripts/Set-Version.ps1`) and automated changelog generation (`scripts/Update-Changelog.ps1`) scripts, while Linux build pipelines and GitHub Actions CI workflows have been refined for reliable cross-platform distribution.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing.

### Added

- Automated changelog generation and update script (`scripts/Update-Changelog.ps1`).
- Modularized `scripts/Set-Version.ps1` featuring interactive tag collision resolution, clean-repository enforcement, and `-CurrentVersion` tagging support.

### Changed

- Rebranded application name to **TextoVista** and updated the Go module (`texto-vista`), binary names, NSIS installer configuration, UI references, and icon assets.
- Upgraded language and runtime requirements to Go 1.27.0 and Node.js >=22.0.0.
- Upgraded desktop application framework to Wails v2.16.0 and updated HTML sanitization handling (`internal/gm-ext/htmlsanitize/htmlsanitize.go`).
- Migrated D2 diagram rendering from `oss.terrastruct.com/d2` to `github.com/d2lang/d2` v0.9.0 and updated `goldmark-d2`.
- Upgraded core Go dependencies (`goldmark` v1.8.6, `chroma/v2` v2.27.0, `gjson` v1.20.0, `regexp2/v2` v2.2.1, and `golang.org/x/*` packages).
- Upgraded frontend dependencies, TypeScript, Vite tooling, and Playwright/Vitest test suites.
- Updated WSL/Linux prerequisite validation (`scripts/verify-wsl-linux.sh`) and application build scripts (`scripts/build-wails-app.sh`) for improved Linux archive packaging.
- Enhanced GitHub Actions workflows and Linux packaging scripts (`scripts/build-wails-app.sh`) for Go 1.27 and consistent `.tar.gz` distribution artifact naming.
- Updated system architecture (`docs/ARCHITECTURE.md`), WSL setup, and Linux packaging guides.

### Fixed

- Resolved Wails v2.16.0 version alignment and Linux packaging configuration across build scripts and GitHub Actions workflows.

## [0.5.2] - 2026-07-05

The application provides configuration options to enable/disable various Markdown syntax enhancements.

This release includes full configuration settings options and includes the ability for the user to select the proportional and monospace font of their choice and set the default font size for each. It has also had some refinements to the internal CSS to improve rendering, though more work needs to be done. Very basic printing support is available as well, though the CSS formatting for printing still needs tweaking.

The application has been compiled and tested for Windows, and the tests and build completed successfully on both WSL2 (Ubuntu-26.04) and within the GitHub CI workflow (to generate a `.tar.gz` package containing the Linux executable). The Go and Wails frameworks are multi-platform and this application could (in theory) be compiled on MacOS as well, but this hasn't been tested at this time.

See the repository's [README](./README.md) for additional information.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage, but development is still ongoing.

### Added

- Manual and automatic document refresh (auto-refresh enabled by default) reloading the currently displayed Markdown file when changes are detected on disk.

### Fixed

- Various security and maintenance fixes.

## [0.5.1] - 2026-06-05

This is the first public release of this Markdown Reader application. It supports most CommonMark specs and GitHub Markdown syntax.

The application provides configuration options to enable/disable various Markdown syntax enhancements.

This release includes full configuration settings options and includes the ability for the user to select the proportional and monospace font of their choice and set the default font size for each. It has also had some refinements to the internal CSS to improve rendering, though more work needs to be done. Very basic printing support is available as well, though the CSS formatting for printing still needs tweaking.

The application has been compiled and tested for Windows, and the tests and build completed successfully on both WSL2 (Ubuntu-26.04) and within the GitHub CI workflow (to generate a `.tar.gz` package containing the Linux executable). The Go and Wails frameworks are multi-platform and this application could (in theory) be compiled on MacOS as well, but this hasn't been tested at this time.

See the repository's [README](./README.md) for additional information.

> [!WARNING]
> This is still **Beta** software! It should be functional for general usage, but development is still ongoing.

### Added

- Initial public release of the Markdown Reader application supporting CommonMark and GitHub Flavored Markdown syntax.
- Full configuration settings options allowing selection of proportional and monospace font families and default font sizes.
- Basic document printing support.
- Windows installer and Linux executable packaging workflows.

[Unreleased]: https://github.com/zmtcreative/TextoVista/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/zmtcreative/TextoVista/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/zmtcreative/TextoVista/compare/v0.5.3...v0.6.0
[0.5.3]: https://github.com/zmtcreative/TextoVista/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/zmtcreative/TextoVista/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/zmtcreative/TextoVista/releases/tag/v0.5.1
