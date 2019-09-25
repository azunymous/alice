package post

import (
	"regexp"
	"strings"
)

var linkedQuoteFirstLine = regexp.MustCompile(`^>>(\d+)[ \t]*`)
var textQuoteFirstLine = regexp.MustCompile(`^>([^>].*)`)

const (
	linkedQuote = `<span class="alc-linked-quote">>$1</span>`
	textQuote   = `<span class="alc-text-quote">$1</span>`
)

func (p Post) parse() string {
	postContent := p.Comment
	var parsedLines []string
	for _, line := range strings.Split(postContent, "\n") {
		line = linkedQuoteFirstLine.ReplaceAllString(line, linkedQuote)
		line = textQuoteFirstLine.ReplaceAllString(line, textQuote)
		parsedLines = append(parsedLines, line)
	}
	return strings.Join(parsedLines, "<br>")
}
