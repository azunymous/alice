package board

import (
	"github.com/alice-ws/alice/data"
	"reflect"
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

func TestStore_AddPost_countDoesNotIncreaseInError(t *testing.T) {
	store := &Store{
		db:    data.NewMemoryDB(),
		count: data.NewMemoryDB(),
	}

	_ = store.db.Set(thread()) // thread in database has number 0
	_ = store.count.Set(data.NewKeyValuePair("/test/:no", "1"))

	_, err := store.AddPost("99", post().with("No", uint64(1)))

	count, _ := store.count.Get("/test/:no")
	if err == nil || count != strconv.Itoa(1) {
		t.Errorf("Expected error and count to be unchanged, got %s, want 1", count)
	}
}

// Utility functions
func post() Post {
	return NewPost(0, time.Unix(0, 0), "Anonymous", "", "Hello World!", "/path/0", "file.png", "")
}

func thread() Thread {
	return NewThread(post(), "A subject")
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
