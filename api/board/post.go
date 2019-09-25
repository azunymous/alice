package post

import (
	"encoding/json"
	"image"
	"path/filepath"
	"sync/atomic"
	"time"
)

type No interface {
	getPostNo() uint64
}

type Post struct {
	No        uint64      `json:"no"`
	Timestamp time.Time   `json:"timestamp"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	Comment   string      `json:"comment"`
	Image     image.Image `json:"image"`
	Filename  string      `json:"filename"`
	Meta      string      `json:"meta"`
	QuotedBy  []uint64    `json:"quoted_by"`
}

func (p Post) Key() string {
	return string(p.No)
}

func (p Post) getPostNo() uint64 {
	return p.No
}

func New(newPostNo uint64, time time.Time, name, email, comment string, image image.Image, filename, meta string) Post {
	return Post{
		No:        newPostNo,
		Timestamp: time,
		Name:      name,
		Email:     email,
		Comment:   comment,
		Image:     image,
		Filename:  filename,
		Meta:      meta,
		QuotedBy:  []uint64{},
	}
}

func addPostQuotedBy(p Post, postQuotingNo uint64) {
	p.QuotedBy = append(p.QuotedBy, postQuotingNo)
}

func (p Post) isValid() bool {
	if len(p.Comment) < 1 && p.Image == nil {
		return false
	}
	if p.Image != nil && filepath.Ext(p.Filename) != ".png" && filepath.Ext(p.Filename) != ".jpeg" && filepath.Ext(p.Filename) != ".jpg" && filepath.Ext(p.Filename) != ".gif" && filepath.Ext(p.Filename) != ".webm" {
		return false
	}

	return true
}

func (p Post) update() Post {
	if len(p.Name) < 1 {
		p.Name = "Anonymous"
	}

	if p.Email == "sage" || p.Email == "noko" || p.Email == "nokosage" {
		p.Meta = p.Email
		p.Email = ""
	}
	return p
}

func (t Post) String() string {
	bytes, _ := json.Marshal(t)
	return "post:" + string(bytes)
}

func (store *Store) AddPost(post Post) (uint64, error) {
	err := store.db.Add(post)
	atomic.AddUint64(&store.count, 1)
	return post.No, err
}
