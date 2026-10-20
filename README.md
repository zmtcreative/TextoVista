# TextoVista

[![License](https://img.shields.io/github/license/zmtcreative/TextoVista)](./LICENSE.md)
![GitHub Release](https://img.shields.io/github/v/release/zmtcreative/TextoVista?sort=semver)

Copyright &copy; 2026 ZMT Creative LLC. All rights reserved.

## About TextoVista

TextoVista is a fast, modern desktop Markdown reader written in Go and TypeScript using the [Wails v2](https://wails.io/) desktop application framework and [Vue 3](https://vuejs.org/).

The reader conforms to the [CommonMark 0.30](https://spec.commonmark.org/) specification powered by [Goldmark](https://github.com/yuin/goldmark) and includes an extensive suite of extensions for technical documentation, diagrams, mathematical expressions, advanced typography, and custom styling.

## Major Features

All features can be toggled and configured via the **File | Settings** dialog (`Ctrl+,`). Settings can be applied for the active session or saved persistently to the application configuration file.

### Core Markdown & Specifications

- **CommonMark Compliance**: Full support for standard CommonMark elements (headings, blockquotes, code blocks, lists, links, images).
- **GitHub Flavored Markdown (GFM)**:
  - Tables with column alignment
  - Task list items with interactive checkboxes
  - Strikethrough (`~~text~~`)
  - Autolinks for URLs and email addresses
- **PHP Markdown Extra**:
  - Definition lists (`Term` followed by `: Definition`)
  - Footnotes (`[^1]` reference and `[^1]: Note`)
  - Abbreviations (`*[HTML]: HyperText Markup Language`)

### Inline Formatting & Typography

- **Text Insertion**: Semantic `<ins>` inserted text markup using `++inserted text++`.
- **Inline Text Highlighting**: Semantic `<mark>` highlighting using `==highlighted text==`.
- **Superscript & Subscript**: Clean mathematical and chemical notations using `^superscript^` and `~subscript~` (with intelligent discrimination preserving `~~strikethrough~~`).
- **Typographic Enhancements (SmartyPants)**:
  - Smart curly quotes (`"quotes"` and `'quotes'`)
  - En-dashes (`--`) and em-dashes (`---`)
  - Proper typographic ellipsis (`...`)
- **Typographic Symbol Shortcuts**:
  - Copyright `(c)` &rarr; &copy;
  - Registered trademark `(r)` &rarr; &reg;
  - Trademark `(tm)` &rarr; &trade;
  - Service mark `(sm)` &rarr; &#8480;
  - Section symbol `(p)` &rarr; &sect;
  - Plus-minus sign `+-` &rarr; &plusmn;
  - Common fractions `(1/2)`, `(1/4)`, `(3/4)` &rarr; &frac12;, &frac14;, &frac34;
  - Directional arrows: `->`, `<-`, `<->`, `=>`, `<=`, `<=>`
- **Emoji Support**: Shortcode and unicode emoji conversions (e.g., `:rocket:`, `:white_check_mark:`).

### Visualizations & Mathematics

- **Mermaid Diagrams & Charts**: Native client-side rendering for flowcharts, sequence diagrams, state diagrams, class diagrams, Gantt charts, and Git graphs via fenced ```` ```mermaid ```` code blocks.
- **D2 Software Architecture Diagrams**: Server-compiled vector diagrams from fenced ```` ```d2 ```` blocks rendered via the Elk layout engine and modern color themes.
- **KaTeX Mathematical Typesetting**: High-fidelity formula rendering for inline equations (`$...$`) and display blocks (`$$...$$`).
- **Code Syntax Highlighting**: Fast, colorful code block syntax highlighting powered by Chroma across dozens of programming and configuration languages.

### Layout & Structural Extensions

- **Alert Callouts & Admonitions**: Styled callout blocks using `> [!NOTE]`, `> [!TIP]`, `> [!IMPORTANT]`, `> [!WARNING]`, and `> [!CAUTION]`. Supports four configurable icon and alias themes:
  - *Strict GFM*: Standard 5 GitHub alert types
  - *GFM with Aliases*: Standard GFM alerts with common name aliases
  - *GFM Plus*: GFM alerts combined with popular Obsidian callouts and custom icons
  - *Obsidian-Style*: Complete Obsidian callout taxonomy and iconography
- **Heading Section Wrapper**: Automatic hierarchical nesting of heading levels within semantic `<section>` elements to maintain logical document structure.
- **Heading Anchor Links**: Automatically generated heading IDs with `#` anchor links for smooth in-document navigation.
- **Fenced DIVs**: Container wrapper syntax using `:::` blocks for targeted CSS class styling.
- **Block Attributes & Bracketed Spans**: Apply custom IDs, classes, and attributes to blocks (`{.custom-class #element-id}`) and inline spans (`[styled text]{.badge}`).
- **Pandoc-Style Fancy Lists**: Advanced ordered lists supporting alphabetical letters (`A.`, `a.`) and Roman numerals (`I.`, `i.`).
- **Image Figure Wrapping**: Enclose images with captions into semantic `<figure>` and `<figcaption>` elements.

### Reader & Application Capabilities

- **Document History Navigation**: In-app navigation history stack with a toolbar **Back** button to easily trace links between Markdown files.
- **Frontmatter Metadata Drawer**: Automatic parsing of YAML, TOML, and JSON frontmatter with an interactive metadata drawer in the toolbar.
- **Live File Watching**: Instant document re-rendering when the active file is modified on disk using background `fsnotify` file watchers.
- **Font Customization**:
  - System font detection across Windows and Linux
  - Independent proportional and monospace font family selectors
  - Granular font size controls
- **Defense-in-Depth Security**:
  - Strict filesystem path validation preventing directory traversal
  - File size safety limits
  - Two-tier HTML sanitization combining Go backend policy filters with frontend DOMPurify validation
  - Safe inline HTML mode with configurable sanitization rules

## Status

> [!NOTE]
>
> TextoVista is currently in **Beta** release.

The application is fully functional for everyday document reading, browsing linked markdown notes, and previewing technical documents containing diagrams, code, and mathematical formulas.

### Current Implementation State

- **Markdown Engine**: Feature-complete with Goldmark 1.8+, D2, Mermaid 12, KaTeX, Chroma syntax highlighting, and an extensive suite of Goldmark extensions.
- **Configuration & Settings**: Fully customizable via the `File | Settings` dialog (`Ctrl+,`) with session and persistent configuration management.
- **Navigation & Usability**: Full document history tracking, smooth internal anchor jumping, and automatic file reloading.
- **Security**: Robust IPC boundary validation and dual-layer HTML sanitization protecting against malicious content.

### Ongoing Development

- **CSS & Themes**: Active styling polish to ensure harmonious visual presentation across all markdown elements, themes, and diagram types.
- **Printing**: Rudimentary printing support is implemented; dedicated print stylesheets and layout adjustments are in progress.
- **Platform Packaging**: Windows installer (NSIS) and Linux packaging workflows are actively maintained and continuously tested.

## Collaboration

This application is currently privately developed and is not accepting external collaborators, feature requests, or pull requests at this time.

The repository is public under the terms of the [Apache 2.0 License](./LICENSE.md). You are welcome to fork the project and modify it as you see fit, provided you adhere to the license and provide appropriate attribution in your `README.md` or `NOTICE.md` file.
