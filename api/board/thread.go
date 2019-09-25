package post

import (
	"encoding/json"
	"github.com/alice-ws/alice/data"
	"sync/atomic"
)

type Store struct {
	db    data.DB
	count uint64
}

func NewStore(db data.DB, key []byte) *Store {
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
		Replies: []Post{},
	}
}

func (t Thread) String() string {
	bytes, _ := json.Marshal(t)
	return "thread:" + string(bytes)
}

func (store *Store) AddThread(thread Thread) (uint64, error) {
	err := store.db.Add(thread)
	atomic.AddUint64(&store.count, 1)
	return thread.No, err
}
