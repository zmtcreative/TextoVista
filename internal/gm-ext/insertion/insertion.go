package insertion

import (
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Kind is the ast.NodeKind for *Insertion AST nodes.
var Kind = gast.NewNodeKind("Insertion")

// Insertion represents an inline text insertion node (<ins>text</ins>).
type Insertion struct {
	gast.BaseInline
}

// Kind returns the node kind.
func (n *Insertion) Kind() gast.NodeKind {
	return Kind
}

// Dump prints the node structure for debugging.
func (n *Insertion) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// NewInsertion creates a new Insertion AST node.
func NewInsertion() *Insertion {
	return &Insertion{}
}

type delimiterProcessor struct{}

var defaultDelimiterProcessor = &delimiterProcessor{}

// IsDelimiter returns true if the character is a '+' delimiter.
func (d *delimiterProcessor) IsDelimiter(b byte) bool {
	return b == '+'
}

// CanOpenCloser returns true if the opener can close the closer delimiter.
func (d *delimiterProcessor) CanOpenCloser(opener *parser.Delimiter, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

// OnMatch creates an Insertion AST node when delimiters match.
func (d *delimiterProcessor) OnMatch(consumes int) gast.Node {
	return NewInsertion()
}

// Parser parses ++insertion++ markdown into an *Insertion AST node.
type Parser struct{}

var defaultParser = &Parser{}

// NewParser returns a new parser.InlineParser for text insertions.
func NewParser() parser.InlineParser {
	return defaultParser
}

// Trigger returns the list of characters that triggers this parser.
func (p *Parser) Trigger() []byte {
	return []byte{'+'}
}

// Parse parses the given block into an inline node.
func (p *Parser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, seg := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 2, defaultDelimiterProcessor)
	if node == nil {
		return nil
	}
	node.Segment = seg.WithStop(seg.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

var _ parser.InlineParser = (*Parser)(nil)

// Renderer renders *Insertion AST nodes into HTML.
type Renderer struct{}

var defaultRenderer = &Renderer{}

// NewRenderer returns a new renderer.NodeRenderer for text insertions.
func NewRenderer() renderer.NodeRenderer {
	return defaultRenderer
}

// RegisterFuncs registers the node renderer for Insertion nodes.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(Kind, r.Render)
}

// Render renders the node into <ins> tags.
func (r *Renderer) Render(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<ins>")
	} else {
		_, _ = w.WriteString("</ins>")
	}
	return gast.WalkContinue, nil
}

// Extender implements goldmark.Extender to support ++insertion++ syntax.
type Extender struct{}

// NewExtender returns a new goldmark.Extender for text insertions.
func NewExtender() goldmark.Extender {
	return &Extender{}
}

// Extend registers the parser and renderer with Goldmark.
func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewParser(), 500),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewRenderer(), 500),
		),
	)
}
