package board

import (
	"encoding/json"
	"errors"
	"github.com/alice-ws/alice/data"
	"sync/atomic"
)

type Store struct {
	db    data.DB
	count uint64
}

func NewStore(db data.DB) *Store {
	if db == nil {
		db = data.NewMemoryDB()
	}

	return &Store{
		db:    db,
		count: 0,
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
	return thread.Post.No, err
}

func (store *Store) GetThread(no string) (Thread, error) {
	threadString, err := store.db.Get(no)
	if err != nil {
		return Thread{}, nil
	}
	return newThreadFrom(threadString)
}

// TODO validate and set post no
func (store *Store) AddPost(threadNo string, post Post) (uint64, error) {
	atomic.AddUint64(&store.count, 1)

	post = post.update(store.count)

	thread, err := store.GetThread(threadNo)

	if err != nil {
		return 0, err
	}

	thread.Replies = append(thread.Replies, post)
	err = store.db.Set(thread)
	return post.No, err
}
