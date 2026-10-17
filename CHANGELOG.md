<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Notes

- Rebrand application name to **TextoVista** and update Go module / binary name to `texto-vista`.
- Ongoing development updates will be collected here until the next tagged release.

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
