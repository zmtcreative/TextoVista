package markdown

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"texto-vista/internal/gm-ext/htmlsanitize"
	utils "texto-vista/internal/utils"

	// "texto-vista/internal/gm-ext/sectionwrapper"
	alertcallouts "github.com/zmtcreative/gm-alert-callouts"
	sectionwrapper "github.com/zmtcreative/gm-sectionwrapper"
	"github.com/d2lang/d2/d2layouts/d2elklayout"
	"github.com/d2lang/d2/d2themes/d2themescatalog"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	figure "github.com/mangoumbrella/goldmark-figure"
	fancylists "github.com/zmtcreative/gm-fancy-lists"

	d2diagrams "github.com/FurqanSoftware/goldmark-d2"
	katex "github.com/kingreatwill/goldmark-katex/v2"

	blockattr "github.com/mdigger/goldmark-attributes"
	bracketedspan "github.com/nemunaire/goldmark-inline-attributes"
	fences "github.com/stefanfritsch/goldmark-fences"
	supersub "github.com/bowman2001/goldmark-supersubscript"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/frontmatter"
	mermaid "go.abhg.dev/goldmark/mermaid"
	texthighlight "github.com/zuern/goldmark-highlight"
	"texto-vista/internal/gm-ext/shortcuts"
)

//go:embed assets/alertcallouts-gfmstrict.icons
var alertCalloutsGFMStrictData string

const (
	ALERT_NOICONS = iota
	ALERT_GFM_STRICT
	ALERT_GFM_WITH_ALIASES
	ALERT_GFM_PLUS
	ALERT_OBSIDIAN
)


var GlobalAttributeFilter = util.NewBytesFilterString(`accesskey,autocapitalize,autofocus,class,contenteditable,dir,draggable,enterkeyhint,hidden,id,inert,inputmode,is,itemid,itemprop,itemref,itemscope,itemtype,lang,part,role,slot,spellcheck,style,tabindex,title,translate`) // nolint:lll
var CodeBlockAttributeFilter = GlobalAttributeFilter.ExtendString(`nolabel,nolable,label,lable`)
var dataPrefix = []byte("data-")

// ConfigProvider interface to avoid circular imports
type ConfigProvider interface {
    // Legacy methods for backward compatibility
    GetApplicationConfig() (useInlineHTML, useSanitize bool)
    GetAlertCalloutConfig() string

    // Application-specific configuration getters
    UseInlineHTML() bool
    UseSanitize() bool
    UseStripH1() bool
    UseFrontmatterTitle() bool

    // Markdown-specific configuration getters
    UseGFM() bool
	UsePHPMDExt() bool
    UseEmoji() bool
    UseMermaid() bool
    UseFigure() bool
    UseAnchor() bool
    UseFences() bool
    UseSections() bool
    UseHighlighting() bool
    UseFancyLists() bool
    UseAttributes() bool
    UseTypographic() bool
	UseAbbreviations() bool
	UseKatex() bool
	UseD2Diagrams() bool
	UseSuperSubScript() bool
	UseInlineHighlighting() bool

    // Alert callouts configuration getters
    UseAlertCallouts() bool
    AlertCalloutStyle() string
}

// CreateGoldmarkInstance creates and configures a new Goldmark instance.
func CreateGoldmarkInstance(configProvider ConfigProvider) goldmark.Markdown {
    // Select alert callout icons based on style
    alertIconID := ALERT_NOICONS
    switch configProvider.AlertCalloutStyle() {
	case "GFMStrict":
		alertIconID = ALERT_GFM_STRICT
    case "GFMWithAliases":
        alertIconID = ALERT_GFM_WITH_ALIASES
    case "GFMPlus":
        alertIconID = ALERT_GFM_PLUS
    case "Obsidian":
        alertIconID = ALERT_OBSIDIAN
    default:
        alertIconID = ALERT_GFM_STRICT // Default to Strict GFM
    }

	// Initialize Goldmark options with bare defaults
	options := []goldmark.Option{
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Automatically generate IDs for headings
		),
		goldmark.WithExtensions(),
		goldmark.WithRendererOptions(),
	}

	// Enable inline HTML rendering if configured
	if configProvider.UseInlineHTML() {
		options = append(options,
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		)
	}

	// Enable GitHub Flavored Markdown (GFM) extensions
	if configProvider.UseGFM() {
		options = append(options,
			goldmark.WithExtensions(
				extension.GFM,
			),
		)
	}

	// Enable PHP Markdown extensions if configured
	if configProvider.UsePHPMDExt() {
		options = append(options,
			goldmark.WithExtensions(
				extension.DefinitionList,
				extension.Footnote,
			),
		)
	}

	// Enable Typographic extensions if configured (for better typography -- fancy quotes and typographic shortcuts)
	if configProvider.UseTypographic() {
		options = append(options,
			goldmark.WithExtensions(
				extension.Typographer,
				shortcuts.NewExtender(),
			),
		)
	}

	// Frontmatter processing is always enabled
	options = append(options,
		goldmark.WithExtensions(
			&frontmatter.Extender{
				Mode: frontmatter.SetMetadata,
			},
		),
	)

	// Enable Mermaid extensions if configured (for Mermaid diagrams/charts)
	if configProvider.UseMermaid() {
		options = append(options,
			goldmark.WithExtensions(
				&mermaid.Extender{},
			),
		)
	}

	// Enable Emoji extensions if configured
	if configProvider.UseEmoji() {
		options = append(options,
			goldmark.WithExtensions(
				emoji.Emoji,
			),
		)
	}

	// Enable Figure extensions if configured (adds <figure> and <figcaption> elements)
	if configProvider.UseFigure() {
		options = append(options,
			goldmark.WithExtensions(
				figure.Figure.WithSkipNoCaption(),
			),
		)
	}

	// Enable Anchor extensions if configured (adds '#' anchor links to headings)
	if configProvider.UseAnchor() {
		options = append(options,
			goldmark.WithExtensions(
				&anchor.Extender{
					Position: anchor.After,
					Texter: anchor.Text("#"),
				},
			),
		)
	}

	// Enable Fences extensions (Fenced Divs) if configured
	if configProvider.UseFences() {
		options = append(options,
			goldmark.WithExtensions(
				&fences.Extender{},
			),
		)
	}

	// Enable section wrapper if configured (nested wrapping of heading sections using <section> elements)
	if configProvider.UseSections() {
		options = append(options,
			goldmark.WithExtensions(
				sectionwrapper.NewSectionWrapper(
					sectionwrapper.WithHeadingClass(true),
				),
			),
		)
	}

	// Enable syntax highlighting if configured
	if configProvider.UseHighlighting() {
		options = append(options,
			goldmark.WithExtensions(
				highlighting.NewHighlighting(
					highlighting.WithStyle("monokailight"),
					highlighting.WithWrapperRenderer(highlightingCustomWrapperRenderer),
					highlighting.WithFormatOptions(
						chromahtml.WithClasses(false),
						chromahtml.PreventSurroundingPre(true), // Let WithWrapperRenderer handle the <pre> tag
						chromahtml.WithAllClasses(false),      // Use all classes for syntax highlighting
						chromahtml.Standalone(true),           // Set to false to prevent a full HTML document
					),
				),
			),
		)
	}

	// Enable Fancy Lists extensions if configured
	if configProvider.UseFancyLists() {
		options = append(options,
			goldmark.WithExtensions(
				&fancylists.FancyListsOptions{},
			),
		)
	}

	// Enable Alert Callouts extensions if configured
	if configProvider.UseAlertCallouts() {
		// Create a new default instance of the alert callouts extender
		acx := alertcallouts.NewAlertCallouts()

		// Add alert callouts based on selected style
		switch alertIconID {
		case ALERT_GFM_STRICT:
			// Use strict GFM icons supplied by the embedded icon set
			acx = alertcallouts.NewAlertCallouts(alertcallouts.UseGFMStrictIcons())

		case ALERT_GFM_WITH_ALIASES:
			// Use standard GFM icons but with aliases for similar alert names (e.g., notes->note)
			acx = alertcallouts.NewAlertCallouts(alertcallouts.UseGFMWithAliasesIcons())

		case ALERT_GFM_PLUS:
			// Use GFM Plus icons (GFM + Obsidian-style callouts but with GFM and custom icons)
			acx = alertcallouts.NewAlertCallouts(alertcallouts.UseGFMPlusIcons())

		case ALERT_OBSIDIAN:
			// Use Obsidian icons (Obsidian-style callouts using Obsidian's icon set)
			acx = alertcallouts.NewAlertCallouts(alertcallouts.UseObsidianIcons())
		}

		// Add the alert callouts extension to the options
		options = append(options,
			goldmark.WithExtensions(acx),
		)

	}

	// Sanitize HTML
	// Note to Self: There is some kind of parsing priority issue between the Sanitizer,
	//               Alert-Callouts and the FancyLists extension (these are all my extensions).
	//               The load order seems to matter here. Something to debug later.
	//
	//               THIS ordering seems to be working properly for now.
    if configProvider.UseSanitize() {
        options = append(options,
            goldmark.WithExtensions(
                htmlsanitize.NewSanitizeHTMLExtension(), // Custom extension to sanitize HTML
            ),
        )
    }

	// Enable Attributes extensions if configured (to use {.myclass #myid} attribute syntax)
	if configProvider.UseAttributes() {
		options = append(options,
			goldmark.WithParserOptions(
				parser.WithAttribute(),     // Enable attributes for nodes
			),
			blockattr.Enable, // Enable block attributes
			bracketedspan.Enable, // Enable bracketed span
		)
	}

	if configProvider.UseKatex() {
		options = append(options,
			goldmark.WithExtensions(
				katex.KaTeX,
			),
		)
	}

	if configProvider.UseD2Diagrams() {
		options = append(options,
			goldmark.WithExtensions(
				&d2diagrams.Extender{
					// Layout: d2dagrelayout.DefaultLayout,
					Layout: d2elklayout.DefaultLayout,
					ThemeID: &d2themescatalog.CoolClassics.ID,
					Sketch: false,
				},
			),
		)
	}

	// Enable Superscript and Subscript if configured.
	// Parsers and renderers run with priority 450 (ahead of GFM strikethrough at 500)
	// so that single tildes (~sub~) are processed first, while double tildes (~~strike~~)
	// fall through to strikethrough.
	if configProvider.UseSuperSubScript() {
		options = append(options,
			goldmark.WithParserOptions(
				parser.WithInlineParsers(
					util.Prioritized(&singleTildeSubscriptParser{inner: supersub.NewSubscriptParser()}, 450),
					util.Prioritized(supersub.NewSuperscriptParser(), 450),
				),
			),
			goldmark.WithRendererOptions(
				renderer.WithNodeRenderers(
					util.Prioritized(supersub.NewSubscriptHTMLRenderer(), 450),
					util.Prioritized(supersub.NewSuperscriptHTMLRenderer(), 450),
				),
			),
		)
	}

	// Enable Inline Text Highlighting (==text==) if configured
	if configProvider.UseInlineHighlighting() {
		options = append(options,
			goldmark.WithExtensions(
				&texthighlight.Extender{},
			),
		)
	}

    return goldmark.New(options...)
}

// singleTildeSubscriptParser wraps the goldmark-supersubscript subscript parser
// so that only single tildes (~sub~) are parsed as subscript, allowing double
// tildes (~~strike~~) to fall through to the GFM strikethrough parser.
type singleTildeSubscriptParser struct {
	inner parser.InlineParser
}

func (s *singleTildeSubscriptParser) Trigger() []byte {
	return s.inner.Trigger()
}

func (s *singleTildeSubscriptParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	count := 0
	for i := 0; i < len(line) && line[i] == '~'; i++ {
		count++
	}
	if count > 1 {
		return nil
	}
	return s.inner.Parse(parent, block, pc)
}

// ConvertMarkdownToHTML converts a byte slice of Markdown content into HTML.
// Returns: the HTML content ([]byte), frontmatter metadata (map[string]any), and any conversion error (error).
func ConvertMarkdownToHTML(mdConverter goldmark.Markdown, markdown []byte, configProvider ConfigProvider) ([]byte, map[string]any, string, error) {
    var buf strings.Builder
	// Get the frontmatter data and place it in meta
	root := mdConverter.Parser().Parse(text.NewReader(markdown))
	doc := root.OwnerDocument()
	meta := doc.Meta()


	// Strip comments from frontmatter before rendering
	// Note to Self: The frontmatter extension handles parsing comments in frontmatter just fine,
	//               but since the original source is passed to the Convert() method, we need to
	//               strip comments in the frontmatter manually -- otherwise, commented
	//               frontmatter will be treated as an ATX header and screw up ExtractH1() and
	//               Convert() calls.
	mdContent, fmType := stripCommentsFromFrontmatter(markdown)

    // Extract the document title from the H1 heading element if present
    // We're doing this before converting, so we can use the Goldmark Parser to find the first '# Title'
    var thisDocumentH1Title, _ = ExtractH1(string(mdContent))

    // Clean up the title by removing extra whitespace and line breaks
    thisDocumentH1Title = strings.ReplaceAll(thisDocumentH1Title, "\n", " ")
    // Replace multiple whitespace characters with a single space using regex
    thisDocumentH1Title = regexp.MustCompile(`\s+`).ReplaceAllString(thisDocumentH1Title, " ")

	// If UseAbbreviations is enabled, get the abbreviation definitions from the markdown
	abbrDefs := make(map[string]string)
	if configProvider.UseAbbreviations() {
		abbrDefs, mdContent = GetMarkdownAbbreviations(mdContent)
		// renderData.Abbreviations = abbrDefs
	}

	// Convert the Markdown content to HTML
	// Note to self: Since we're altering the mdContent before rendering,
	//               we need use the Convert() method, we can't use the 'root'
	//               variable with the Renderer() function here, since the source
	//               now doesn't match the AST in 'root' anymore.
	err := mdConverter.Convert(mdContent, &buf)
    if err != nil {
        return nil, nil, "", err
    }
    html := buf.String()
	meta, _ = utils.NormalizeMapKeys(meta) // Normalize keys to lowercase

	// Add custom keys AFTER nomalizing -- only actual frontmatter keys should be lowercased
	meta["__FMTYPE__"] = fmType // Add frontmatter type to metadata
	if len(abbrDefs) > 0 {
		meta["__ABBR__"] = abbrDefs
	}

    return []byte(html), meta, thisDocumentH1Title, nil
}

// ExtractH1 finds the first H1 heading from Markdown source.
// func ExtractH1(md string) (string, []byte, error) {
func ExtractH1(md string) (string, error) {
    source := []byte(md)
    mdParser := goldmark.DefaultParser()
    reader := text.NewReader(source)
    doc := mdParser.Parse(reader)

    var h1Node *ast.Heading
    ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
        if !entering || h1Node != nil {
            return ast.WalkContinue, nil
        }
        if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
            h1Node = heading
            return ast.WalkStop, nil
        }
        return ast.WalkContinue, nil
    })

    if h1Node == nil {
        return "", nil
    }

    title := ExtractTextContent(h1Node, source)
    if strings.TrimSpace(title) == "" {
        return title, nil
    }

    // modifiedSource := RemoveNodeFromSource(source, h1Node)
    return title, nil
}

func highlightingCustomWrapperRenderer(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
	if entering {
		lang, _ := c.Language()
		lang = getLanguageByAlias(lang)
        // Add language class to the <pre> tag
		fmt.Fprintf(w, `<pre language="%s"`, lang)
		if c.Attributes() != nil {
			renderCodeBlockAttributes(w, c, CodeBlockAttributeFilter)
		}
		fmt.Fprintf(w, `>`)
		// Add language class to the <code> tag
		fmt.Fprintf(w, `<code class="chroma" language="%s">`, lang)
	} else {
		_, _ = w.WriteString(`</code></pre>`)
	}
}

func renderCodeBlockAttributes(w util.BufWriter, c highlighting.CodeBlockContext, filter util.BytesFilter) {
	for _, attr := range c.Attributes().All() {
		if filter != nil && !filter.Contains(attr.Name) {
			if !bytes.HasPrefix(attr.Name, dataPrefix) {
				continue
			}
		}
		_, _ = w.WriteString(" ")
		_, _ = w.Write(attr.Name)
		_, _ = w.WriteString(`="`)
		var value []byte
		switch typed := attr.Value.(type) {
			case []byte:
				value = typed
			case string:
				value = util.StringToReadOnlyBytes(typed)
		}
		_, _ = w.Write(util.EscapeHTML(value))
		_ = w.WriteByte('"')
	}
}

func ExtractTextContent(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			segment := textNode.Segment
			value := segment.Value(source)
			buf.Write(value)
			if textNode.SoftLineBreak() {
				buf.WriteByte(' ')
			}
		} else {
			buf.WriteString(ExtractTextContent(child, source))
		}
	}
	return buf.String()
}

func RemoveNodeFromSource(source []byte, node ast.Node) string {
    lines := node.Lines()
    if lines.Len() == 0 {
        return string(source)
    }

    // Get the full range from first line start to last line end
    firstLine := lines.At(0)
    lastLine := lines.At(lines.Len() - 1)

    // Find the actual start of the line (including any preceding whitespace and # markers)
    start := firstLine.Start
    // Go backwards to find the beginning of the line
    for start > 0 && source[start-1] != '\n' && source[start-1] != '\r' {
        start--
    }

    end := lastLine.Stop

    // Extend removal range to include trailing newline if present
    if end < len(source) && source[end] == '\n' {
        end++
    } else if end < len(source)-1 && source[end] == '\r' && source[end+1] == '\n' {
        end += 2
    }

    // Remove the node's segment from the source
    return string(source[:start]) + string(source[end:])
}

func InitAlertCalloutsIcons(icondata string) map[string]string {
	ai := make(map[string]string)

	// Parse the embedded alert callouts data
	lines := strings.Split(icondata, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if it's an alias definition (contains ->)
		if strings.Contains(line, "->") {
			parts := strings.SplitN(line, "->", 2)
			if len(parts) == 2 {
				alias := strings.TrimSpace(parts[0])
				primary := strings.TrimSpace(parts[1])
				// Set alias to reference the primary icon (will be set after core icons are loaded)
				// Forward-reference aliases (primary not yet seen) are resolved by the second pass below.
				if svg, exists := ai[primary]; exists {
					ai[alias] = svg
				}
			}
			continue
		}

		// Parse core icon definition (key|svg)
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			svg := strings.TrimSpace(parts[1])
			ai[key] = svg
		}
	}

	// Second pass to handle any aliases that couldn't be resolved in first pass
	lines = strings.Split(icondata, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "->") {
			parts := strings.SplitN(line, "->", 2)
			if len(parts) == 2 {
				alias := strings.TrimSpace(parts[0])
				primary := strings.TrimSpace(parts[1])
				if svg, exists := ai[primary]; exists {
					ai[alias] = svg
				}
			}
		}
	}

	return ai
}

type CodeLanguage struct {
	Name string
	Aliases []string // Aliases for the language
}

var CodeLanguages = []CodeLanguage{
	{Name: "actionscript", Aliases: []string{"actionscript", "as", "as3", "actionscript3"}},
	{Name: "awk", Aliases: []string{"awk", "gawk", "nawk", "mawk"}},
	{Name: "bash", Aliases: []string{"bash", "bashrc", "ksh", "sh", "shell", "zsh", "zshrc"}},
	{Name: "basic", Aliases: []string{"basic", "bas", "qbasic"}},
	{Name: "batchfile", Aliases: []string{"batchfile", "bat", "cmd", "dos", "dosbatch", "windowsbatch", "winbatch"}},
	{Name: "c", Aliases: []string{"c", "c89", "c90", "c99", "c11", "h", "idc"}},
	{Name: "c++", Aliases: []string{"c++", "cpp", "cxx", "cc", "c++11", "c++14", "c++17", "c++20", "h++", "hpp", "hxx"}},
	{Name: "c#", Aliases: []string{"c#", "csharp", "cs"}},
	{Name: "clojure", Aliases: []string{"clojure", "clj"}},
	{Name: "coffeescript", Aliases: []string{"coffeescript", "coffee", "coffee-script", "cjs"}},
	{Name: "diff", Aliases: []string{"diff", "patch", "gitdiff", "gdiff", "udiff"}},
	{Name: "f#", Aliases: []string{"f#", "fsharp", "fs"}},
	{Name: "fortran", Aliases: []string{"fortran", "f77", "f90", "f95", "f03", "f08"}},
	{Name: "gnuplot", Aliases: []string{"gnuplot", "gp", "plot"}},
	{Name: "go", Aliases: []string{"go", "golang"}},
	{Name: "groovy", Aliases: []string{"groovy", "gradle", "gvy", "gy"}},
	{Name: "haskell", Aliases: []string{"haskell", "hs"}},
	{Name: "html", Aliases: []string{"html", "htm", "xhtml"}},
	{Name: "ini", Aliases: []string{"ini", "dosini", "winini", "inf"}},
	{Name: "javascript", Aliases: []string{"javascript", "js", "jsm"}},
	{Name: "julia", Aliases: []string{"julia", "jl"}},
	{Name: "kotlin", Aliases: []string{"kotlin", "kt"}},
	{Name: "lisp", Aliases: []string{"lisp", "cl", "common-lisp"}},
	{Name: "makefile", Aliases: []string{"makefile", "mf", "mk", "mak", "make"}},
	{Name: "markdown", Aliases: []string{"markdown", "md"}},
	{Name: "perl", Aliases: []string{"perl", "pl", "pm", "pod"}},
	{Name: "postscript", Aliases: []string{"postscript", "ps", "eps"}},
	{Name: "powershell", Aliases: []string{"powershell", "ps1", "psm1", "psd1", "pwsh", "posh"}},
	{Name: "python", Aliases: []string{"python", "py", "py3", "py2", "python2", "python3"}},
	{Name: "rexx", Aliases: []string{"rexx", "rex", "arexx", "rx"}},
	{Name: "ruby", Aliases: []string{"ruby", "rb", "rbw", "rbx", "rake", "gemspec", "gemfile"}},
	{Name: "rust", Aliases: []string{"rust", "rs"}},
	{Name: "typescript", Aliases: []string{"typescript", "ts", "tsx", "tsm"}},
}

func getLanguageByAlias(language []byte) []byte {
	for _, lang := range CodeLanguages {
		if lang.Name == string(language) || slices.Contains(lang.Aliases, string(language)) {
			return []byte(lang.Name)
		}
	}
	return language
}

// stripCommentsFromFrontmatter removes comment lines (starting with #) from frontmatter
func stripCommentsFromFrontmatter(mdContent []byte) ([]byte, string) {
    content := string(mdContent)
    lines := strings.Split(content, "\n")

    var result strings.Builder
    var frontmatterDelimiter string
    inFrontmatter := false
    frontmatterStarted := false
	fmType := ""

    for i, line := range lines {
        trimmedLine := strings.TrimSpace(line)

        // Check if we're still in the preamble (only blank lines allowed before frontmatter)
        if !frontmatterStarted && !inFrontmatter {
            // If this is a blank line, keep it and continue
            if trimmedLine == "" {
                result.WriteString(line)
                if i < len(lines)-1 {
                    result.WriteString("\n")
                }
                continue
            }

            // Check if this line starts frontmatter
            if strings.HasPrefix(trimmedLine, "---") && len(trimmedLine) >= 3 && strings.Trim(trimmedLine, "-") == "" {
                frontmatterDelimiter = trimmedLine
                frontmatterStarted = true
                inFrontmatter = true
				fmType = "YAML"
                result.WriteString(line)
                if i < len(lines)-1 {
                    result.WriteString("\n")
                }
                continue
            } else if strings.HasPrefix(trimmedLine, "+++") && len(trimmedLine) >= 3 && strings.Trim(trimmedLine, "+") == "" {
                frontmatterDelimiter = trimmedLine
                frontmatterStarted = true
                inFrontmatter = true
                fmType = "TOML"
                result.WriteString(line)
                if i < len(lines)-1 {
                    result.WriteString("\n")
                }
                continue
            } else {
                // No frontmatter detected, this is regular content
                // Add all remaining content as-is
                result.WriteString(strings.Join(lines[i:], "\n"))
                break
            }
        }

        // We are inside frontmatter
        if inFrontmatter {
            // Check if this is the closing delimiter
            if trimmedLine == frontmatterDelimiter {
                inFrontmatter = false
                result.WriteString(line)
                if i < len(lines)-1 {
                    result.WriteString("\n")
                }
                continue
            }

            // Skip lines that start with # (comments)
            if strings.HasPrefix(trimmedLine, "#") {
                continue
            }

            // Keep non-comment lines
            result.WriteString(line)
            if i < len(lines)-1 {
                result.WriteString("\n")
            }
        } else {
            // We're past the frontmatter, add everything as-is
            result.WriteString(line)
            if i < len(lines)-1 {
                result.WriteString("\n")
            }
        }
    }

    return []byte(result.String()), fmType
}