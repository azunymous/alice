package threads

import (
	"encoding/json"
	"time"
)

type BoardResponse struct {
	Status string `json:"status"`
	No     string `json:"no"`
	Thread Thread `json:"thread"`
	Type   string `json:"type"`
}

type Thread struct {
	Post    `json:"post"`
	Subject string `json:"subject"`
	Replies []Post `json:"replies"`
}

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

type Segment struct {
	Format  []string `json:"format"`
	Segment string   `json:"segment"`
}

func (r BoardResponse) AsJSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (t Thread) AsJSON() string {
	b, _ := json.Marshal(t)
	return string(b)
}

func (p Post) AsJSON() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ThreadFromJSON(thread string) Thread {
	t := &Thread{}
	_ = json.Unmarshal([]byte(thread), t)
	return *t
}

func PostFromJSON(post string) Post {
	p := &Post{}
	_ = json.Unmarshal([]byte(post), p)
	return *p
}

func post() Post {
	return Post{
		No:              0,
		Timestamp:       time.Unix(0, 0),
		Name:            "Anonymous",
		Email:           "",
		Comment:         "Hello World!",
		CommentSegments: make([]Segment, 0),
		Image:           "group/0.png",
		Filename:        "0.png",
		Meta:            "",
		QuotedBy:        make([]uint64, 0),
	}
}

func thread() Thread {
	return Thread{Post: post(), Subject: "a subject", Replies: []Post{}}
}
