package threads

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"gopkg.in/h2non/gentleman.v2/plugins/multipart"
	"image"
	"image/png"
	"log"
	"strconv"
	"strings"
	"time"
)

const boardID = "/obj/"

const redisAddr = "localhost:6379"

const (
	empty     = -1
	undefined = 0

	adding = 1
	added  = 2

	prepare = 4
	getting = 5
	got     = 6

	assert = 7
)

type Controller struct {
	redis                   *redis.Client
	threadUnderModification Thread
	threadFromDatabase      string
	threads                 threadList
	state                   int
	formFields              multipart.FormData
}

type boardResponse struct {
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

func (r boardResponse) AsJSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (t Thread) AsJSON() string {
	b, _ := json.Marshal(t)
	return string(b)
}

func (t Thread) HasSameCoreFields(t2 Thread) bool {
	if t.No != t2.No || t.Comment != t2.Comment {
		return false
	}
	return true
}

func FromJSON(thread string) Thread {
	t := &Thread{}
	_ = json.Unmarshal([]byte(thread), t)
	return *t
}

func Operation() *Controller {
	return &Controller{redis: redisClient()}
}

// TODO Change to be parallelisable
func (tm *Controller) ClearRedis() *Controller {
	tm.redis.FlushAll()
	tm.state = empty
	return tm
}

func (tm *Controller) Add() *Controller {
	tm.state = adding
	return tm
}

func (tm *Controller) Thread(threadNo ...string) *Controller {
	switch tm.state {
	case adding:
		tm.threadUnderModification = thread()
	case getting:
		get := tm.redis.Get(threadNo[0])
		if get.Err() != nil {
			panic("getting thread from redis error: " + get.Err().Error())
		}
		tm.threadFromDatabase = get.Val()
	}
	return tm
}

func (tm *Controller) And() *Controller {
	tm.finalisedThread()
	tm.state = adding
	return tm
}

func (tm *Controller) AnotherThread() *Controller {
	tm.threadUnderModification = thread()
	tm.threadUnderModification.No = 1
	tm.threadUnderModification.Timestamp = tm.threadUnderModification.Timestamp.Add(1 * time.Nanosecond)
	return tm
}

func (tm *Controller) WithNo(no uint64) *Controller {
	tm.threadUnderModification.No = no
	return tm
}

// TODO separate into more generalised function
func (tm *Controller) ToRedis() *Controller {
	if tm.state == adding {
		tm.finalisedThread()
		for _, t := range tm.threads {
			tm.redis.Set(boardID+":no", t.No+1, 0)
			tm.redis.ZAdd(boardID, redis.Z{Score: float64(t.Timestamp.UnixNano()), Member: t.No})
			tm.redis.Set(strconv.FormatUint(t.No, 10), t.AsJSON(), 0)
		}
	} else {
		panic("no thread was set up to be added to Redis")
	}
	return tm
}

func (tm *Controller) Fields() multipart.FormData {
	return tm.formFields
}

func (tm *Controller) ExpectedResponse(threadNo ...int) string {
	switch tm.state {
	case added:
		response := boardResponse{
			Status: "SUCCESS",
			No:     "0",
			Thread: tm.threads[0],
			Type:   "THREAD",
		}
		if len(threadNo) > 0 {
			response.No = strconv.FormatUint(tm.threads[threadNo[0]].No, 10)
			response.Thread = tm.threads[threadNo[0]]
		}
		return response.AsJSON()
	case prepare:
		return `{"status":"SUCCESS","username":"","error":"","token":""}`
	}
	panic("nothing to expect")
}

func (tm *Controller) ExpectedArray() string {
	switch tm.state {
	case added:
		return "[" + strings.Join(tm.threads.asJSON(), ",") + "]"
	}
	panic("nothing to expect")
}

func (tm *Controller) Check() *Controller {
	tm.state = assert
	return tm
}

func (tm *Controller) IfEqualToExpectedThread() *Controller {
	if !FromJSON(tm.threadFromDatabase).HasSameCoreFields(tm.threads[0]) {
		log.Fatalf("thread not equal in core fields to expected. got %s expected %s", tm.threadFromDatabase, tm.threads[0].AsJSON())
	}
	return tm
}

func (tm *Controller) finalisedThread() Thread {
	tm.threads = append(tm.threads, tm.threadUnderModification)
	defer func() { tm.threadUnderModification = Thread{} }()
	tm.state = added
	return tm.threadUnderModification
}

// PrepareToPostThread initializes chain to create the required structs for posting a new thread including multipart form field and an image
func (tm *Controller) PrepareToPostThread() *Controller {
	tm.state = prepare
	return tm
}

func (tm *Controller) WithFields() *Controller {
	_ = png.Encode(new(bytes.Buffer), image.NewRGBA(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{X: 100, Y: 100}}))
	const comment = "Hello World!"
	file := []multipart.FormFile{
		{
			Name:   "image",
			Reader: new(bytes.Buffer),
		},
	}

	fields := multipart.FormData{
		Data:  map[string]multipart.Values{"comment": {comment}},
		Files: file,
	}

	tm.threads = append(tm.threads, Thread{
		Post: Post{
			No:      0,
			Comment: comment,
		},
	})
	tm.formFields = fields
	return tm
}

func (tm *Controller) Get() *Controller {
	tm.state = getting
	return tm
}

func redisClient() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: redisAddr})
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
	return Thread{Post: post(), Subject: "a subject"}
}

type threadList []Thread

func (tl threadList) asJSON() []string {
	var listAsJSON []string

	// Reverse order
	for i := len(tl) - 1; i >= 0; i-- {
		listAsJSON = append(listAsJSON, tl[i].AsJSON())
	}
	return listAsJSON
}
