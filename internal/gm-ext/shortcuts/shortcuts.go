package shortcuts

import (
	"bytes"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	copyReplacement   = []byte("&copy;")   // ©
	regReplacement    = []byte("&reg;")    // ®
	tradeReplacement  = []byte("&trade;")  // ™
	sectReplacement   = []byte("&sect;")   // §
	smReplacement     = []byte("&#8480;")  // ℠
	plusmnReplacement = []byte("&plusmn;") // ±
	frac12Replacement = []byte("&frac12;") // ½
	frac14Replacement = []byte("&frac14;") // ¼
	frac34Replacement = []byte("&frac34;") // ¾
	rarrReplacement   = []byte("&rarr;")   // →
	larrReplacement   = []byte("&larr;")   // ←
	harrReplacement   = []byte("&harr;")   // ↔
	rArrReplacement   = []byte("&rArr;")   // ⇒
	lArrReplacement   = []byte("&lArr;")   // ⇐
	hArrReplacement   = []byte("&hArr;")   // ⇔
)

type typographicShortcutsParser struct{}

var defaultTypographicShortcutsParser = &typographicShortcutsParser{}

// NewParser returns a new parser.InlineParser that converts typographic shortcuts.
func NewParser() parser.InlineParser {
	return defaultTypographicShortcutsParser
}

func (p *typographicShortcutsParser) Trigger() []byte {
	return []byte{'(', '+', '<', '-', '='}
}

func (p *typographicShortcutsParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
	if len(line) == 0 {
		return nil
	}

	switch line[0] {
	case '(':
		return parseParenthesized(block, line)
	case '+':
		return parsePlusMinus(block, line)
	case '<', '-', '=':
		return parseArrow(block, line)
	default:
		return nil
	}
}

func parseParenthesized(block text.Reader, line []byte) gast.Node {
	if len(line) < 3 || line[0] != '(' {
		return nil
	}

	// (c) or (C) -> © (&copy;)
	if line[2] == ')' && (line[1] == 'c' || line[1] == 'C') {
		node := gast.NewString(copyReplacement)
		node.SetCode(true)
		block.Advance(3)
		return node
	}

	// (r) or (R) -> ® (&reg;)
	if line[2] == ')' && (line[1] == 'r' || line[1] == 'R') {
		node := gast.NewString(regReplacement)
		node.SetCode(true)
		block.Advance(3)
		return node
	}

	// (p) or (P) -> § (&sect;)
	if line[2] == ')' && (line[1] == 'p' || line[1] == 'P') {
		node := gast.NewString(sectReplacement)
		node.SetCode(true)
		block.Advance(3)
		return node
	}

	// (tm), (TM), (Tm), or (tM) -> ™ (&trade;)
	if len(line) >= 4 && line[3] == ')' &&
		(line[1] == 't' || line[1] == 'T') &&
		(line[2] == 'm' || line[2] == 'M') {
		node := gast.NewString(tradeReplacement)
		node.SetCode(true)
		block.Advance(4)
		return node
	}

	// (sm), (SM), (Sm), or (sM) -> ℠ (&#8480;)
	if len(line) >= 4 && line[3] == ')' &&
		(line[1] == 's' || line[1] == 'S') &&
		(line[2] == 'm' || line[2] == 'M') {
		node := gast.NewString(smReplacement)
		node.SetCode(true)
		block.Advance(4)
		return node
	}

	return nil
}

func parsePlusMinus(block text.Reader, line []byte) gast.Node {
	if len(line) < 2 || line[0] != '+' || line[1] != '-' {
		return nil
	}

	node := gast.NewString(plusmnReplacement)
	node.SetCode(true)
	block.Advance(2)
	return node
}

func isAlphanumericOrSlash(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		r == '/'
}

func parseArrow(block text.Reader, line []byte) gast.Node {
	switch line[0] {
	case '<':
		if bytes.HasPrefix(line, []byte("<=>")) {
			node := gast.NewString(hArrReplacement)
			node.SetCode(true)
			block.Advance(3)
			return node
		}
		if bytes.HasPrefix(line, []byte("<-->")) {
			node := gast.NewString(harrReplacement)
			node.SetCode(true)
			block.Advance(4)
			return node
		}
		if bytes.HasPrefix(line, []byte("<->")) {
			node := gast.NewString(harrReplacement)
			node.SetCode(true)
			block.Advance(3)
			return node
		}
		if bytes.HasPrefix(line, []byte("<==")) {
			node := gast.NewString(lArrReplacement)
			node.SetCode(true)
			block.Advance(3)
			return node
		}
		if bytes.HasPrefix(line, []byte("<--")) {
			node := gast.NewString(larrReplacement)
			node.SetCode(true)
			block.Advance(3)
			return node
		}
		if bytes.HasPrefix(line, []byte("<-")) {
			node := gast.NewString(larrReplacement)
			node.SetCode(true)
			block.Advance(2)
			return node
		}
	case '-':
		if bytes.HasPrefix(line, []byte("-->")) {
			node := gast.NewString(rarrReplacement)
			node.SetCode(true)
			block.Advance(3)
			return node
		}
		if bytes.HasPrefix(line, []byte("->")) {
			node := gast.NewString(rarrReplacement)
			node.SetCode(true)
			block.Advance(2)
			return node
		}
	case '=':
		if bytes.HasPrefix(line, []byte("==>")) {
			node := gast.NewString(rArrReplacement)
			node.SetCode(true)
			block.Advance(3)
			return node
		}
		if bytes.HasPrefix(line, []byte("=>")) {
			node := gast.NewString(rArrReplacement)
			node.SetCode(true)
			block.Advance(2)
			return node
		}
	}
	return nil
}

type shortcutsExtender struct{}

// NewExtender returns a new goldmark.Extender for typographic shortcuts.
func NewExtender() goldmark.Extender {
	return &shortcutsExtender{}
}

func (e *shortcutsExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(defaultTypographicShortcutsParser, 9999),
		),
		parser.WithASTTransformers(
			util.Prioritized(defaultFractionsASTTransformer, 9999),
		),
	)
}

type fractionsASTTransformer struct{}

var defaultFractionsASTTransformer = &fractionsASTTransformer{}

func (f *fractionsASTTransformer) Transform(doc *gast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	_ = gast.Walk(doc, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}
		if n.Kind() == gast.KindCodeSpan || n.Kind() == gast.KindFencedCodeBlock || n.Kind() == gast.KindCodeBlock {
			return gast.WalkSkipChildren, nil
		}
		if textNode, ok := n.(*gast.Text); ok {
			transformFractionsInText(textNode, source)
		}
		return gast.WalkContinue, nil
	})
}

func transformFractionsInText(textNode *gast.Text, source []byte) {
	seg := textNode.Segment
	content := seg.Value(source)
	if len(content) < 3 {
		return
	}

	for i := 0; i <= len(content)-3; i++ {
		if content[i+1] != '/' {
			continue
		}

		var rep []byte
		if content[i] == '1' && content[i+2] == '2' {
			rep = frac12Replacement
		} else if content[i] == '1' && content[i+2] == '4' {
			rep = frac14Replacement
		} else if content[i] == '3' && content[i+2] == '4' {
			rep = frac34Replacement
		} else {
			continue
		}

		// Boundary check before
		var prev byte
		if i > 0 {
			prev = content[i-1]
		} else if seg.Start > 0 && seg.Start <= len(source) {
			prev = source[seg.Start-1]
		}
		if isAlphanumericOrSlash(rune(prev)) {
			continue
		}

		// Boundary check after
		var next byte
		if i+3 < len(content) {
			next = content[i+3]
		} else if seg.Stop < len(source) {
			next = source[seg.Stop]
		}
		if isAlphanumericOrSlash(rune(next)) {
			continue
		}

		parent := textNode.Parent()
		if parent == nil {
			return
		}

		if i > 0 {
			prefixSeg := text.NewSegment(seg.Start, seg.Start+i)
			prefixNode := gast.NewTextSegment(prefixSeg)
			parent.InsertBefore(parent, textNode, prefixNode)
		}

		fracNode := gast.NewString(rep)
		fracNode.SetCode(true)
		parent.InsertBefore(parent, textNode, fracNode)

		if i+3 < len(content) {
			suffixSeg := text.NewSegment(seg.Start+i+3, seg.Stop)
			suffixNode := gast.NewTextSegment(suffixSeg)
			suffixNode.SetSoftLineBreak(textNode.SoftLineBreak())
			suffixNode.SetHardLineBreak(textNode.HardLineBreak())
			parent.InsertBefore(parent, textNode, suffixNode)
			parent.RemoveChild(parent, textNode)
			transformFractionsInText(suffixNode, source)
			return
		}

		parent.RemoveChild(parent, textNode)
		return
	}
}
