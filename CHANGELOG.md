<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

This update refines the Linux build and packaging pipeline alongside corresponding developer documentation. Corrections have been applied to the Wails build script, GitHub Actions CI workflow, and architectural guides to ensure reliable Linux distribution packaging and accurate system documentation.

### CHANGES

- Fix: Corrected Linux packaging configuration across build scripts and CI workflow definitions.
- CI/CD: Updated `.github/workflows/build.yml` to align Linux build and artifact packaging steps.
- Scripts: Refined `scripts/build-wails-app.sh` for improved Linux application building and archive generation.
- Docs: Updated `docs/ARCHITECTURE.md` and `docs/_DEV_Linux-Packaging-tar-gz.md` to reflect current Linux packaging procedures and architectural details.

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
