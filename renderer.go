package main

import (
	"bytes"

	"golang.org/x/tools/godoc"

	"github.com/russross/blackfriday"
)

type Renderer struct {
	*blackfriday.Html
}

const commonHtmlFlags = 0 |
	blackfriday.HTML_USE_XHTML |
	blackfriday.HTML_USE_SMARTYPANTS |
	blackfriday.HTML_SMARTYPANTS_FRACTIONS |
	blackfriday.HTML_SMARTYPANTS_DASHES |
	blackfriday.HTML_SMARTYPANTS_LATEX_DASHES

const commonExtensions = 0 |
	blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
	blackfriday.EXTENSION_TABLES |
	blackfriday.EXTENSION_FENCED_CODE |
	blackfriday.EXTENSION_AUTOLINK |
	blackfriday.EXTENSION_STRIKETHROUGH |
	blackfriday.EXTENSION_SPACE_HEADERS |
	blackfriday.EXTENSION_HEADER_IDS |
	blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
	blackfriday.EXTENSION_DEFINITION_LISTS

var (
	bfHtmlRenderer = blackfriday.HtmlRenderer(commonHtmlFlags, "", "")
	renderer       = &Renderer{Html: bfHtmlRenderer.(*blackfriday.Html)}
)

func (options *Renderer) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	switch lang {
	case "go":
		out.WriteString("<pre>")
		godoc.FormatText(out, text, 1, true, "", nil)
		out.WriteString("</pre>")
	case "shell":
		out.WriteString("<div class='shell'>")
		bfHtmlRenderer.BlockCode(out, text, lang)
		out.WriteString("</div>")
	case "output":
		out.WriteString("<div class='output'>")
		bfHtmlRenderer.BlockCode(out, text, lang)
		out.WriteString("</div>")
	default:
		bfHtmlRenderer.BlockCode(out, text, lang)
	}
}
