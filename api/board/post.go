package board

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"time"
)

type Post struct {
	No        uint64    `json:"no"`
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Comment   string    `json:"comment"`
	Image     string    `json:"image"`
	Filename  string    `json:"filename"`
	Meta      string    `json:"meta"`
	QuotedBy  []uint64  `json:"quoted_by"`
}

func (p Post) Key() string {
	return strconv.FormatUint(p.No, 10)
}

func (p Post) getPostNo() uint64 {
	return p.No
}

func NewPost(newPostNo uint64, time time.Time, name, email, comment string, image string, filename, meta string) Post {
	return Post{
		No:        newPostNo,
		Timestamp: time,
		Name:      name,
		Email:     email,
		Comment:   comment,
		Image:     image,
		Filename:  filename,
		Meta:      meta,
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

func addPostQuotedBy(p Post, postQuotingNo uint64) {
	p.QuotedBy = append(p.QuotedBy, postQuotingNo)
}

func (p Post) isValid() bool {
	if len(p.Comment) < 1 && p.Image == "" {
		return false
	}
	if p.Image != "" && filepath.Ext(p.Filename) != ".png" && filepath.Ext(p.Filename) != ".jpeg" && filepath.Ext(p.Filename) != ".jpg" && filepath.Ext(p.Filename) != ".gif" && filepath.Ext(p.Filename) != ".webm" {
		return false
	}

	return true
}

func (p Post) update(postCount uint64) Post {
	p.No = postCount - 1

	if len(p.Name) < 1 {
		p.Name = "Anonymous"
	}

	if p.Email == "sage" || p.Email == "noko" || p.Email == "nokosage" {
		p.Meta = p.Email
		p.Email = ""
	}
	return p
}

func (p Post) String() string {
	bytes, _ := json.Marshal(p)
	return "post:" + string(bytes)
}
