package board

import (
	"encoding/json"
	"errors"
	"github.com/alice-ws/alice/data"
	"log"
	"strconv"
)

type Store struct {
	ID      string
	db      data.KeyValueDB
	count   data.KeyValueDB
	threads data.OrderedDB
}

func NewStore(ID string, db data.KeyValueDB, threads data.OrderedDB) *Store {
	if db == nil {
		db = data.NewMemoryDB()
		threads = data.NewMemoryDB()
	}

	store := &Store{
		ID:      ID,
		db:      db,
		count:   db,
		threads: threads,
	}

	// Set the the board count to 0 if the key does not exist.
	if _, err := db.Get(boardCountKey(store)); err != nil {
		err := db.Set(data.NewKeyValuePair(boardCountKey(store), "0"))
		if err != nil {
			log.Printf("Could not create thread DB: %v", err)
		}
	}

	return store
}

// Returns key for board count that is stored in the DB
func boardCountKey(store *Store) string {
	return store.ID + ":no"
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

// Returns the index and post of the reply with the given post no.
// Returns -1 if the no is not found in the replies.
func (t Thread) getReplyWithPostNo(no uint64) (index int, p Post) {
	for i, p := range t.Replies {
		if p.No == no {
			return i, p
		}
	}
	return -1, Post{}
}

func (store *Store) AddThread(thread Thread) (uint64, error) {
	currentNumberOfPosts := store.incrementAndGet()

	// Ignore transformations as the thread is empty. (No cross thread transformations for now)
	thread.Post, _ = thread.update(currentNumberOfPosts)
	err := store.db.Set(thread)
	err = store.threads.SetOrdered(data.NewKeyValuePair(store.ID, strconv.FormatUint(thread.No, 10)), int(thread.Timestamp.Unix()))
	return thread.Post.No, err
}

func (store *Store) GetAllThreads() ([]Thread, error) {
	var threads []Thread

	for _, t := range store.threads.GetAllOrderedByScore(store.ID) {
		threadString, err := store.db.Get(t)
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

func (store *Store) AddPost(threadNo string, post Post) (uint64, error) {
	thread, err := store.GetThread(threadNo)

	if err != nil {
		return 0, err
	}

	currentNumberOfPosts := store.incrementAndGet()
	post, threadTransformations := post.update(currentNumberOfPosts)

	thread.Replies = append(thread.Replies, post)
	for _, transformation := range threadTransformations {
		thread = transformation(thread)
	}
	err = store.db.Set(thread)
	err = store.threads.SetOrdered(data.NewKeyValuePair(store.ID, strconv.FormatUint(thread.No, 10)), int(post.Timestamp.Unix()))
	return post.No, err
}

func (store *Store) incrementAndGet() uint64 {
	currentCount, err := store.count.Increment(boardCountKey(store))
	if err != nil {
		log.Printf("Cannot increment post count.")
	}
	return uint64(currentCount)
}
