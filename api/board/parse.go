package board

import (
	"regexp"
	"strconv"
	"strings"
)

type format struct {
	regex                  *regexp.Regexp
	class                  string
	transformationProvider func(f format, line string, p Post) func(t Thread) Thread
}

type Transform func(t Thread) Thread

func getFormats() []format {
	return []format{
		{
			regex: regexp.MustCompile(`^>([^>].*)`),
			class: "quote",
		},
		{
			regex:                  regexp.MustCompile(`^>>(\d+)[ \t]*`),
			class:                  "noQuote",
			transformationProvider: addPostReplyToQuotedPost,
		},
	}
}

// One line with either a format or not
type Segment struct {
	Format  []string `json:"format"`
	Segment string   `json:"segment"`
}

func (p Post) parse() (Post, []Transform) {
	post := p
	postContent := post.Comment
	var segments []Segment
	var transformations []Transform
	for _, line := range strings.Split(postContent, "\n") {
		addingSegment := Segment{[]string{}, line}
		for _, f := range getFormats() {
			find := f.regex.FindString(line)
			if find != "" {
				addingSegment = Segment{[]string{f.class}, line}
				if f.transformationProvider != nil {
					transformations = append(transformations, f.transformationProvider(f, line, p))
				}
			}
		}
		segments = append(segments, addingSegment)
	}

	post.CommentSegments = segments
	return post, transformations
}

func addPostReplyToQuotedPost(f format, line string, p Post) func(t Thread) Thread {
	return func(t Thread) Thread {
		submatches := f.regex.FindStringSubmatch(line)
		if len(submatches[1]) > 0 {
			quotedPostNo, _ := strconv.ParseUint(submatches[1], 10, 0)
			index, transformedPost := t.getReplyWithPostNo(quotedPostNo)
			if quotedPostNo == t.No {
				t.Post = t.Post.quotedBy(p.No)
			} else if index > -1 {
				t.Replies[index] = transformedPost.quotedBy(p.No)
			}
		}
		return t
	}
}
