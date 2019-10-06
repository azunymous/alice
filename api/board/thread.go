package board

import (
	"encoding/json"
	"errors"
	"github.com/alice-ws/alice/data"
	"strconv"
	"sync/atomic"
)

type Store struct {
	db      data.DB
	count   uint64
	threads []uint64
}

func NewStore(db data.DB) *Store {
	if db == nil {
		db = data.NewMemoryDB()
	}

	return &Store{
		db:      db,
		count:   0,
		threads: []uint64{},
	}
}

type Thread struct {
	Post    `json:"post"`
	Subject string `json:"subject"`
	Replies []Post `json:"replies"`
}

func NewThread(post Post, subject string) Thread {
	return Thread{
		Post:    post,
		Subject: subject,
		Replies: make([]Post, 0),
	}
}

func newThreadFrom(mjson string) (Thread, error) {
	var t Thread
	err := json.Unmarshal([]byte(mjson), &t)
	if err != nil {
		return Thread{}, errors.New("cannot parse json" + err.Error())
	}
	return t, nil
}

func (t Thread) String() string {
	bytes, _ := json.Marshal(t)
	return string(bytes)
}

func (store *Store) AddThread(thread Thread) (uint64, error) {
	atomic.AddUint64(&store.count, 1)

	thread.Post = thread.update(store.count)
	err := store.db.Set(thread)
	store.threads = append(store.threads, thread.Post.No)
	return thread.Post.No, err
}

func (store *Store) GetAllThreads() ([]Thread, error) {
	var threads []Thread

	for _, t := range store.threads {
		threadString, err := store.db.Get(strconv.FormatUint(t, 10))
		thread, err := newThreadFrom(threadString)
		if err != nil {
			return []Thread{}, errors.New("error getting threads")
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

func (store *Store) GetThread(no string) (Thread, error) {
	threadString, err := store.db.Get(no)
	if err != nil {
		return Thread{}, errors.New("no such thread found")
	}
	return newThreadFrom(threadString)
}

// TODO validate and set post no
func (store *Store) AddPost(threadNo string, post Post) (uint64, error) {
	thread, err := store.GetThread(threadNo)

	if err != nil {
		return 0, err
	}

	atomic.AddUint64(&store.count, 1)
	post = post.update(store.count)

	thread.Replies = append(thread.Replies, post)
	err = store.db.Set(thread)
	return post.No, err
}
