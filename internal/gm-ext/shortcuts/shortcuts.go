package shortcuts

import (
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	copyReplacement  = []byte("&copy;")
	regReplacement   = []byte("&reg;")
	tradeReplacement = []byte("&trade;")
)

type typographicShortcutsParser struct{}

var defaultTypographicShortcutsParser = &typographicShortcutsParser{}

// NewParser returns a new parser.InlineParser that converts (c), (tm), and (r) shortcuts.
func NewParser() parser.InlineParser {
	return defaultTypographicShortcutsParser
}

func (p *typographicShortcutsParser) Trigger() []byte {
	return []byte{'('}
}

func (p *typographicShortcutsParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
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

	// (tm), (TM), (Tm), or (tM) -> ™ (&trade;)
	if len(line) >= 4 && line[3] == ')' &&
		(line[1] == 't' || line[1] == 'T') &&
		(line[2] == 'm' || line[2] == 'M') {
		node := gast.NewString(tradeReplacement)
		node.SetCode(true)
		block.Advance(4)
		return node
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
	)
}
