package board

import (
	"regexp"
	"strings"
)

type format struct {
	regex *regexp.Regexp
	class string
}

func getFormats() []format {
	return []format{
		{
			regex: regexp.MustCompile(`^>([^>].*)`),
			class: "quote",
		},
		{
			regex: regexp.MustCompile(`^>>(\d+)[ \t]*`),
			class: "noQuote",
		},
	}
}

// One line with either a format or not
type Segment struct {
	format  []string
	segment string
}

func (p Post) parse() []Segment {
	postContent := p.Comment
	var segments []Segment
	for _, line := range strings.Split(postContent, "\n") {
		addingSegment := Segment{[]string{}, line}
		for _, f := range getFormats() {
			find := f.regex.FindString(line)
			if find != "" {
				addingSegment = Segment{[]string{f.class}, line}
			}
		}
		segments = append(segments, addingSegment)
	}
	return segments
}
