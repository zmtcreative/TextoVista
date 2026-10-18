---
Title: TextoVista Markdown Feature Showcase
Author: TextoVista Team
Date: 2026-10-18
# DocType: TechDoc
Tags: [markdown, commonmark, gfm, katex, mermaid, d2]
---

<!-- markdownlint-disable MD004 MD018 MD025 MD033 MD034 MD036 MD042 MD046 MD049 MD051 -->

# TextoVista Markdown Feature Showcase

Welcome to the **TextoVista** feature demonstration document. This file showcases the complete suite of CommonMark standards, GitHub Flavored Markdown (GFM) extensions, and integrated Goldmark plugins supported by the application.

## Table of Contents

- [Frontmatter & Document Metadata](#frontmatter--document-metadata)
- [CommonMark Typography & Formatting](#commonmark-typography--formatting)
- [Lists & Task Lists](#lists--task-lists)
- [Tables & Tabular Data](#tables--tabular-data)
- [Code Blocks & Syntax Highlighting](#code-blocks--syntax-highlighting)
- [Alert Callouts & Admonitions](#alert-callouts--admonitions)
- [Math Expressions (KaTeX)](#math-expressions-katex)
- [Diagrams & Visualizations (Mermaid & D2)](#diagrams--visualizations-mermaid--d2)
- [Extended Syntax & Attributes](#extended-syntax--attributes)
- [HTML & Security Sanitization](#html--security-sanitization)
- [Media, Figures & Links](#media-figures--links)

---

## Frontmatter & Document Metadata

TextoVista automatically parses YAML, TOML, and JSON frontmatter blocks positioned at the very start of a document.

Metadata attributes such as `Title`, `Author`, `Date`, and custom tags are extracted into structured document properties. In the TextoVista desktop interface, click the **Frontmatter** button in the top toolbar to toggle the metadata drawer and inspect rendered frontmatter fields.

```yaml
---
Title: TextoVista Markdown Feature Showcase
Author: TextoVista Team
Date: 2026-10-18
DocType: TechDoc
Tags: [markdown, commonmark, gfm, katex, mermaid, d2]
---
```

---

## CommonMark Typography & Formatting

### Headings

Headings from Level 1 through Level 6 are supported with automatic heading ID generation for internal anchor linking:

# Heading Level 1
## Heading Level 2
### Heading Level 3
#### Heading Level 4
##### Heading Level 5
###### Heading Level 6

### Paragraphs & Text Emphasis

Markdown supports fine-grained inline emphasis and typographic styling:

- **Bold text** using `**bold**` or `__bold__`
- *Italic text* using `*italic*` or `_italic_`
- ***Bold and italic text*** using `***bold and italic***`
- ~~Strikethrough text~~ using `~~strikethrough~~`
- `Inline code spans` using `` `inline code` ``
- Hard line breaks using two trailing spaces or a backslash (`\`) at the end of a line.

### Blockquotes

Blockquotes can be simple single-line citations, multi-paragraph notes, or deeply nested structures:

> **Basic Blockquote:** Fast, elegant, and distraction-free Markdown viewing built on Go, Wails, and modern Web standards.

Multi-paragraph and nested blockquotes:

> First paragraph of a comprehensive architectural note.
>
> > **Nested Quote:** Deeply nested citations inherit harmonious theme-aware border styling and indented layouts.
>
> Resuming the outer blockquote context.

### Typographic Extensions (SmartyPants)

The typographic extension automatically transforms straight ASCII punctuation into publication-quality typographic glyphs:

- Straight double quotes (`"Hello, World!"`) become curly quotes: "Hello, World!"
- Straight single quotes (`'TextoVista'`) become curly quotes: 'TextoVista'
- En-dash (`--`) becomes an en-dash: 2024--2026
- Em-dash (`---`) becomes an em-dash: Go---the systems language
- Ellipsis (`...`) becomes an ellipsis glyph: Loading...

---

## Lists & Task Lists

### Unordered Lists

Unordered lists support multiple bullet styles (`-`, `*`, `+`) and arbitrary nested levels:

- Systems Layer
  - Go Runtime
    - Concurrency with goroutines
    - Channel synchronization
  - Wails v2 Native IPC Bridge
- Presentation Layer
  - Vue 3 Reactive Components
  - SCSS Design System & Themes

### Ordered Lists

Standard sequential numbered lists:

1. Initial Environment Audit & Sanity Check
2. Security Vulnerability Remediation
3. Race Condition & Concurrency Resolution
4. Test Verification across Backend and Frontend

### Pandoc-Style Fancy Lists

When the `gm-fancy-lists` extension is active, ordered lists can use letters, roman numerals, and auto-numbering markers:

A. Architecture Phase
   i. Frontend Layout Design
   ii. Backend Data Pipeline
   #. Subsystem Finalization
B. Implementation Phase
   a. Unit & Integration Testing
   b. Performance Benchmarks
C. Deployment Phase

### GFM Task Lists

Interactive-style task lists with completion status checkboxes:

- [x] Integrate Goldmark Core parser
- [x] Configure GFM Extensions (Tables, Strikethrough, Autolinks, Task Lists)
- [x] Implement Mermaid diagramming renderer
- [x] Implement D2 diagram compilation
- [x] Add KaTeX mathematical typesetting
- [ ] Implement live collaborative sync

### Definition Lists

PHP Markdown Extra definition lists for glossaries and term specifications:

CommonMark
: A strongly specified, highly compatible specification of Markdown syntax.

Goldmark
: An extensible, high-performance Markdown parser implemented in Go conforming to CommonMark 0.30.

Wails
: An application framework that allows developers to write desktop applications using Go and web frontend technologies.

---

## Tables & Tabular Data

GitHub Flavored Markdown (GFM) tables support customizable column alignments and rich inline styling:

| Feature / Extension | Engine / Library | Purpose | Status |
| :--- | :---: | :---: | ---: |
| **CommonMark Parser** | `goldmark` | Core Markdown Parsing | Supported |
| **Syntax Highlighting** | `chroma` | Code Block Colorization | Supported |
| **Math Formatting** | `katex` | LaTeX Mathematical Formulas | Supported |
| **Diagrams (Flowchart)** | `mermaid.js` | Dynamic Diagram Generation | Supported |
| **Declarative Diagrams** | `d2` | Elegant Software Diagrams | Supported |
| **Admonitions** | `gm-alertcallouts` | GitHub & Obsidian Callouts | Supported |

---

## Code Blocks & Syntax Highlighting

TextoVista uses Chroma for fast, colorful, theme-aware syntax highlighting across dozens of languages.

### Go

```go
package main

import (
	"context"
	"fmt"
)

type Document struct {
	Title string `json:"title"`
	Lines int    `json:"lines"`
}

func main() {
	doc := Document{Title: "Feature Showcase", Lines: 42}
	fmt.Printf("Loaded %s with %d lines.\n", doc.Title, doc.Lines)
}
```

### TypeScript / Vue

```typescript
import { ref, computed } from 'vue';

interface DocumentState {
  title: string;
  wordCount: number;
  isReady: boolean;
}

export function useDocument() {
  const document = ref<DocumentState>({ title: '', wordCount: 0, isReady: false });
  const isAvailable = computed(() => document.value.isReady && document.value.wordCount > 0);
  return { document, isAvailable };
}
```

### Python

```python
from dataclasses import dataclass
from typing import List

@dataclass
class Benchmark:
    name: str
    scores: List[float]

    @property
    def average(self) -> float:
        return sum(self.scores) / len(self.scores) if self.scores else 0.0

bench = Benchmark(name="RenderEngine", scores=[98.2, 99.1, 99.7])
print(f"Benchmark {bench.name}: {bench.average:.2f}%")
```

### JSON Configuration

```json
{
  "application": {
    "name": "TextoVista",
    "version": "0.5.3",
    "autoRefresh": true,
    "theme": "light"
  },
  "typography": {
    "fontFamily": "Inter, sans-serif",
    "fontSize": 16.0
  }
}
```

---

## Alert Callouts & Admonitions

TextoVista features comprehensive support for GitHub-style and Obsidian-style alert callouts (admonitions) with dedicated icons, themes, and collapsible folding.

### Primary Admonition Types

> [!NOTE]
> **Informational Note:** Highlights useful contextual details or background information.

> [!TIP]
> **Pro Tip:** Offers actionable advice, best practices, or productivity shortcuts.

> [!IMPORTANT]
> **Crucial Guidance:** Highlights essential requirements, steps, or prerequisites.

> [!WARNING]
> **Warning Notice:** Cautions users about breaking changes, pitfalls, or deprecations.

> [!CAUTION]
> **Caution Alert:** Advises against dangerous actions that could lead to data loss or security issues.

> [!QUESTION]
> **Help & Inquiries:** Used for FAQs, questions, and clarification requests.

> [!BUG]
> **Known Issue:** Pinpoints open bugs, edge cases, or tracking tickets.

> [!SUCCESS]
> **Operation Success:** Marks completed tasks, passing tests, and confirmed achievements.

> [!QUOTE]
> **Citation Callout:** Highlights authoritative quotes, research excerpts, and literature references.

> [!TODO]
> **Action Item:** Designates pending tasks, checklists, and next steps.

### Collapsible / Folding Callouts

Callouts can be configured to fold or collapse using `+` (open by default) or `-` (collapsed by default):

> [!TIP]- Collapsed Tip (Click to expand)
> This callout is rendered using native `<details>` and `<summary>` elements and is closed by default.

> [!IMPORTANT]+ Open by Default (Explicit)
> This callout is explicitly opened by default while allowing readers to collapse it.

### Custom Titles & Icon-less Callouts

You can override the title of any callout, or omit the icon:

> [!TIP] Keyboard Navigation Shortcut
> Press <kbd>Ctrl</kbd> + <kbd>O</kbd> to open a file dialog, or <kbd>Ctrl</kbd> + <kbd>,</kbd> to access Settings.

> [!NoIcon] Clean Text Callout
> This callout omits the leading icon badge while preserving the thematic accent bar and background card.

### Callout Aliases Reference

All common alias names map automatically to their respective primary callout styling:

| Alias Marker | Target Category | Default Theme Accent |
| :--- | :--- | :--- |
| `[!INFO]`, `[!NOTES]`, `[!INFORMATION]` | Note | Accent Blue |
| `[!HINT]`, `[!HINTS]`, `[!TIPS]` | Tip | Emerald Green |
| `[!WARN]`, `[!WARNINGS]`, `[!ATTENTION]` | Warning | Amber Gold |
| `[!DANGER]`, `[!ERROR]`, `[!ERRORS]` | Caution | Ruby Red |
| `[!FAIL]`, `[!FAILURE]`, `[!MISSING]` | Failure | Crimson Red |
| `[!FAQ]`, `[!FAQS]`, `[!HELP]` | Question | Purple |
| `[!DONE]`, `[!CHECK]` | Success | Bright Green |
| `[!ABSTRACT]`, `[!OVERVIEW]`, `[!SUMMARY]` | Summary | Teal |
| `[!TASK]`, `[!TASKLIST]`, `[!CHECKLIST]` | Todo | Indigo |

---

## Math Expressions (KaTeX)

Mathematical equations and scientific formulas are rendered using KaTeX.

### Inline Math

Inline math expressions integrate seamlessly inside running text:
- Euler's identity: $e^{i\pi} + 1 = 0$
- Einstein's mass-energy formula: $E = mc^2$
- Quadratic formula root: $x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$
- Parenthetical notation syntax: \( \lim_{x \to 0} \frac{\sin x}{x} = 1 \)

### Display / Block Math

Complex multiline formulas and calculus equations can be centered in standalone display blocks:

$$
\oint_C \mathbf{B} \cdot d\mathbf{l} = \mu_0 I_{\text{enc}} + \mu_0 \varepsilon_0 \frac{d\Phi_E}{dt}
$$

Cauchy's integral formula:

$$
f(a) = \frac{1}{2\pi i} \oint_\gamma \frac{f(z)}{z - a} \, dz
$$

Normal distribution probability density:

$$
P(x) = \frac{1}{\sigma \sqrt{2\pi}} \exp\left( -\frac{(x - \mu)^2}{2\sigma^2} \right)
$$

---

## Diagrams & Visualizations (Mermaid & D2)

### Mermaid Flowchart

```mermaid
flowchart TD
    A[Markdown File] --> B[Goldmark Parser]
    B --> C{Render Pipeline}
    C -->|HTML & Metadata| D[Wails IPC Bridge]
    D --> E[Vue Frontend]
    E --> F[Interactive Document View]
```

### Mermaid Sequence Diagram

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Frontend as Vue 3 UI
    participant Backend as Go Backend
    participant FS as Local Filesystem

    User->>Frontend: Click "Open File"
    Frontend->>Backend: OpenFileMenuHandler()
    Backend->>FS: Prompt Native File Dialog
    FS-->>Backend: Return Selected Path
    Backend->>FS: Validate (Size, Path, Non-Binary)
    Backend->>Backend: ConvertMarkdownToHTML()
    Backend-->>Frontend: Emit 'markdown-rendered'
    Frontend-->>User: Render Interactive Document
```

### D2 Diagram

TextoVista integrates the modern declarative diagramming language **D2** (`goldmark-d2`), compiling directly into crisp, scale-independent SVG:

```d2
direction: right

Client: TextoVista Desktop UI {
  style.fill: "#e3f2fd"
}

Host: Go / Wails Application {
  style.fill: "#e8f5e9"
  Parser: Goldmark Pipeline
  Watcher: fsnotify Daemon
}

Client -> Host: Wails IPC Calls
Host.Parser -> Client: HTML + Metadata Payload
Host.Watcher -> Host: Document Change Notification
```

---

## Extended Syntax & Attributes

### Custom Attributes & Bracketed Spans

The attribute extension allows assigning CSS classes, identifiers, and inline styles directly to Markdown blocks or bracketed spans:

[Highlighted accent text span]{style="background-color: #fff3cd; color: #856404; padding: 2px 6px; border-radius: 4px; font-weight: 600;"}

Custom block attributes:
{.custom-container style="border-left: 3px solid #28a745; padding-left: 12px; font-style: italic;"}
This paragraph demonstrates custom block attributes applied dynamically to the rendered document.

### Fenced Divs

Fenced divs provide semantic container grouping using three or more colons (`:::`):

::: info-box
**Semantic Fenced Div Container**
This entire box is enclosed within an HTML `<div>` container styled with custom classes.
:::

### Footnotes

Footnotes are referenced in running text and automatically organized at the bottom of the document:

TextoVista is powered by the Go programming language[^golang] and utilizes Wails v2[^wails] to drive its native UI window.

[^golang]: Go is an open-source programming language supported by Google.
[^wails]: Wails is a desktop application framework pairing Go with modern web interfaces.

### Abbreviations

The abbreviation extension creates `<abbr>` tags with hover tooltips for declared acronyms:

The IPC bridge communicates between the DOM and the Go backend without external server latency.

*[IPC]: Inter-Process Communication
*[DOM]: Document Object Model

### Emojis

Standard GitHub emoji shortcodes are converted into native unicode emojis:

:rocket: Rocket &nbsp;|&nbsp; :sparkles: Sparkles &nbsp;|&nbsp; :fire: Fire &nbsp;|&nbsp; :bulb: Idea &nbsp;|&nbsp; :white_check_mark: Complete &nbsp;|&nbsp; :tada: Celebration

---

## HTML & Security Sanitization

### Supported Inline HTML

Safe standard HTML elements are supported for specialized layout needs:

- Keyboard key visualization: <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>P</kbd>
- Highlighted text: <mark>Highlighted text marker</mark>
- Scientific sub/superscripts: H<sub>2</sub>O and x<sup>2</sup> + y<sup>2</sup> = z<sup>2</sup>
- Native HTML collapsible details:
  <details>
    <summary>Click to view technical specifications</summary>
    <p>TextoVista renders Markdown with synchronous sub-millisecond conversion and reactive DOM updates.</p>
  </details>

### Security & Sanitizer Protections

TextoVista enforces robust defense-in-depth sanitization on both the Go backend (`htmlsanitize`) and the Vue frontend (`DOMPurify`):

- **Forbidden Dangerous Tags:** Elements like `<script>`, `<iframe>`, `<object>`, `<embed>`, `<applet>`, `<base>`, `<meta>`, and `<form>` are strictly stripped.
- **Event Handler Neutralization:** All inline event handlers (`onload`, `onerror`, `onclick`, etc.) are removed.
- **URI Scheme Restrictions:** Attribute URLs are limited exclusively to safe protocols (`http`, `https`, `mailto`, `tel`, `file`). Unsafe schemes like `javascript:` or malicious `data:` URIs are neutralized.

---

## Media, Figures & Links

### Links

- Internal Document Navigation: [Return to Table of Contents](#table-of-contents)
- Relative Project Links: [View README.md](../../README.md)
- External Safe Links: [CommonMark Specification](https://spec.commonmark.org/)

### Figures & Captions

The figure extension converts captioned images into semantic `<figure>` and `<figcaption>` elements:

![TextoVista Vector Icon](../icon/svg/texto-vista-icon.svg "TextoVista Application Icon")

### Images & Embedded Data URIs

Both relative local paths and compact embedded base64 data URIs are supported:

![TextoVista Design](danger-square.svg "Sample Vector Asset")

Simple, sample image (URL source):

![Sample Image](https://dummyimage.com/64x64/cccccc/000000.png&text=Sample "Sample Image")

Sample compact 16x16 PNG Data URI:

![Status Indicator](data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAB1UExURf8RmJkKWwEAAAAAAAgVGhxIVlq856t2f9JNR9xBOb1jZWqs0VnR9/wdCXivyXWyzb9gYdFNSFTX/3axy/8cB1/K7lzN84Wgts1STsBgYKd8hfQmFJqKmMNcXVnD7puIlqN/ipIJV9pEPL1iZN8+NCRGUpMJWNLEbKgAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAZdEVYdFNvZnR3YXJlAFBhaW50Lk5FVCA1LjEuMTITAUd0AAAAuGVYSWZJSSoACAAAAAUAGgEFAAEAAABKAAAAGwEFAAEAAABSAAAAKAEDAAEAAAACAAAAMQECABEAAABaAAAAaYcEAAEAAABsAAAAAAAAAGAAAAABAAAAYAAAAAEAAABQYWludC5ORVQgNS4xLjEyAAADAACQBwAEAAAAMDIzMAGgAwABAAAAAQAAAAWgBAABAAAAlgAAAAAAAAACAAEAAgAEAAAAUjk4AAIABwAEAAAAMDEwMAAAAADZp5qVybcLXwAAAH9JREFUKFNtz9cOwjAQRNH1FtNCgCF0SKH9/ycib8AQiXmwdM+DZRP9W2AREQ7fVjMz/UhgjaPxZDp7S+piXi6Wq9gLa0TautpEZSIS80ZZFTDJsN3tD/iF4+mMIVzS6cBaA2haALVf2rk0V+8uPcTlds/tYvaw3ETP/nO5B3sB9fkItcjyZOUAAAAASUVORK5CYII= "Inline Base64 PNG Sample")

## Embedded Images

![MSU Logo][image1]
MSU Logo (_using a WebP image embedded in the Markdown document using Base64_)

<a id="docbottom"></a>

[image1]: data:image/webp;base64,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

---
*End of TextoVista Feature Showcase Document*
