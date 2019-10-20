package board

import (
	"github.com/alice-ws/alice/data"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestPost_validate(t *testing.T) {

	tests := []struct {
		name      string
		inputPost Post
		want      bool
	}{
		{
			name:      "passes validation for text only valid post",
			inputPost: post(),
			want:      true,
		},
		{
			name:      "passes validation for png",
			inputPost: post().with("Filename", "file.png"),
			want:      true,
		},
		{
			name:      "passes validation for jpg",
			inputPost: post().with("Filename", "file.jpg"),
			want:      true,
		},
		{
			name:      "passes validation for jpeg",
			inputPost: post().with("Filename", "file.jpeg"),
			want:      true,
		},
		{
			name:      "passes validation for gif",
			inputPost: post().with("Filename", "file.gif"),
			want:      true,
		},
		{
			name:      "passes validation for webm",
			inputPost: post().with("Filename", "file.webm"),
			want:      true,
		},
		{
			name:      "fails validation for exe (non png, jpg, jpeg, gif or webm)",
			inputPost: post().with("Image", "/path/0").with("Filename", "file.exe"),
			want:      false,
		},
		{
			name:      "fails validation for zip (non png, jpg, jpeg, gif or webm)",
			inputPost: post().with("Image", "/path/0").with("Filename", "file.zip"),
			want:      false,
		},
		{
			name:      "fails validation for no file ext (non png, jpg, jpeg, gif or webm)",
			inputPost: post().with("Image", "/path/0").with("Filename", "file"),
			want:      false,
		},
		{
			name:      "fails validation for no file name (non png, jpg, jpeg, gif or webm)",
			inputPost: post().with("Image", "/path/0").with("Filename", ""),
			want:      false,
		},
		{
			name:      "fails validation for double filename (non png, jpg, jpeg, gif or webm)",
			inputPost: post().with("Image", "/path/0").with("Filename", "file.jpg.exe"),
			want:      false,
		},
		{
			name:      "passes validation for image with no message",
			inputPost: post().with("Image", "/path/0").with("Filename", "file.png").with("Comment", ""),
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.inputPost.IsValid(); got != tt.want {
				t.Errorf("isValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPost_update(t *testing.T) {
	tests := []struct {
		name      string
		inputPost Post
		want      Post
	}{
		{
			name:      "add 'Anonymous' to name field if name is missing",
			inputPost: post().with("Name", ""),
			want:      post().with("Name", "Anonymous"),
		},
		{
			name:      "move sage from email field to meta field",
			inputPost: post().with("Email", "sage"),
			want:      post().with("Email", "").with("Meta", "sage"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.inputPost.update(1); !equalToIgnoringTime(got, tt.want) {
				t.Errorf("update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_AddPost(t *testing.T) {
	type fields struct {
		db    data.KeyValueDB
		count data.KeyValueDB
	}
	type args struct {
		post Post
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(input Post, returnValue uint64, store Store) bool
		wantErr bool
	}{
		{
			name: "first post is created with incremented post no and post is added to thread",
			fields: fields{
				db:    data.NewMemoryDB(),
				count: data.NewMemoryDB(),
			},
			args: args{
				post: post(),
			},
			want:    countIsIncrementedForPost,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:    tt.fields.db,
				count: tt.fields.count,
			}
			_ = store.db.Set(thread()) // thread in database has number 0

			got, err := store.AddPost("0", tt.args.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want(tt.args.post, got, *store) {
				t.Errorf("AddPost() got = %v, want %v", got, runtime.FuncForPC(reflect.ValueOf(tt.want).Pointer()).Name())
			}
		})
	}
}

func TestStore_AddPost_countDoesNotIncreaseInError(t *testing.T) {
	store := &Store{
		db:    data.NewMemoryDB(),
		count: data.NewMemoryDB(),
	}
	_ = store.db.Set(thread()) // thread in database has number 0

	_, err := store.AddPost("99", post().with("No", uint64(1)))

	count, _ := store.count.Get("/test/")
	if err == nil || count != strconv.Itoa(1) {
		t.Errorf("Expected error and count to be unchanged, got %d, want 1", store.count)
	}
}

// Utility functions
func post() Post {
	return NewPost(0, time.Unix(0, 0), "Anonymous", "", "Hello World!", "/path/0", "file.png", "")
}

func (p Post) with(fieldName string, value interface{}) Post {
	reflect.ValueOf(&p).Elem().FieldByName(fieldName).Set(reflect.ValueOf(value))
	return p
}

func equalToIgnoringTime(p, p2 Post) bool {
	p.Timestamp = time.Unix(0, 0)
	p2.Timestamp = time.Unix(0, 0)
	return reflect.DeepEqual(p, p2)
}

func countIsIncrementedForPost(input Post, _ uint64, store Store) bool {
	get, err := store.db.Get("0")

	expectedPost := input
	expectedPost.No = 4

	got, _ := newThreadFrom(get)
	count, _ := store.count.Get("/test/")
	if err == nil && count == strconv.Itoa(5) && equalToIgnoringTime(expectedPost, got.Replies[0]) {
		return true
	}

	println("Count was not incremented and/or DB does not contain the post")
	return false
}
