package board

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"time"
)

type Post struct {
	No              uint64    `json:"no"`
	Timestamp       time.Time `json:"timestamp"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Comment         string    `json:"comment"`
	CommentSegments []Segment `json:"comment_segments"`
	Image           string    `json:"image"`
	Filename        string    `json:"filename"`
	Meta            string    `json:"meta"`
	QuotedBy        []uint64  `json:"quoted_by"`
}

func (p Post) Key() string {
	return strconv.FormatUint(p.No, 10)
}

func (p Post) getPostNo() uint64 {
	return p.No
}

func NewPost(newPostNo uint64, time time.Time, name, email, comment string, image string, filename, meta string) Post {
	return Post{
		No:              newPostNo,
		Timestamp:       time,
		Name:            name,
		Email:           email,
		Comment:         comment,
		CommentSegments: make([]Segment, 0),
		Image:           image,
		Filename:        filename,
		Meta:            meta,
		QuotedBy:        make([]uint64, 0),
	}
}

func CreatePost(name, email, comment string) Post {
	return Post{
		Timestamp: time.Time{},
		Name:      name,
		Email:     email,
		Comment:   comment,
		QuotedBy:  make([]uint64, 0),
	}
}

func newPostFrom(mjson string) (Post, error) {
	var p Post
	err := json.Unmarshal([]byte(mjson), &p)
	if err != nil {
		return Post{}, errors.New("cannot parse json" + err.Error())
	}
	return p, nil
}

func (p Post) quotedBy(postQuotingNo uint64) Post {
	p.QuotedBy = append(p.QuotedBy, postQuotingNo)
	return p
}

func (p Post) IsValid() bool {
	if len(p.Comment) < 1 && p.Image == "" {
		return false
	}
	if p.Image != "" && filepath.Ext(p.Filename) != ".png" && filepath.Ext(p.Filename) != ".jpeg" && filepath.Ext(p.Filename) != ".jpg" && filepath.Ext(p.Filename) != ".gif" && filepath.Ext(p.Filename) != ".webm" {
		return false
	}

	return true
}

func (p Post) update(postCount uint64) (Post, []Transform) {
	post := p
	post.No = postCount - 1

	if len(post.Name) < 1 {
		post.Name = "Anonymous"
	}

	if post.Email == "sage" || post.Email == "noko" || post.Email == "nokosage" {
		post.Meta = post.Email
		post.Email = ""
	}

	post.Timestamp = time.Now()

	post, threadTransformations := post.parse()
	return post, threadTransformations
}

func (p Post) String() string {
	bytes, _ := json.Marshal(p)
	return "post:" + string(bytes)
}
