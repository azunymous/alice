package board

import (
	"github.com/alice-ws/alice/data"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestStore_AddThread(t *testing.T) {
	type fields struct {
		db    data.DB
		index uint64
	}
	type args struct {
		thread Thread
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(input Thread, returnValue uint64, store Store) bool
		wantErr bool
	}{
		{
			name: "first thread is created with incrementing post no and thread is added with post no 0",
			fields: fields{
				db:    data.NewMemoryDB(),
				index: 0,
			},
			args: args{
				thread: thread(),
			},
			want:    incrementedForFirstThread,
			wantErr: false,
		},
		{
			name: "thread is stored in a readable format and can be converted back to a thread afterwards",
			fields: fields{
				db:    data.NewMemoryDB(),
				index: 1,
			},
			args: args{
				thread: thread(),
			},
			want:    threadIsReadable,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:    tt.fields.db,
				count: tt.fields.index,
			}
			got, err := store.AddThread(tt.args.thread)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddThread() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want(tt.args.thread, got, *store) {
				t.Errorf("AddThread() got = %v, want %v", got, runtime.FuncForPC(reflect.ValueOf(tt.want).Pointer()).Name())
			}
		})
	}
}

func TestThread_String(t *testing.T) {
	thread := thread()

	gotThreadString := thread.String()
	expectedThreadString := `{"post":{"no":0,"timestamp":"1970-01-01T01:00:00+01:00","name":"Anonymous","email":"","comment":"Hello World!","image":"/path/0","filename":"file.png","meta":"","quoted_by":[]},"subject":"A subject","replies":[]}`
	if gotThreadString != expectedThreadString {
		t.Errorf("Thread string was %s, want %s", gotThreadString, expectedThreadString)
	}
}

func incrementedForFirstThread(input Thread, returnValue uint64, store Store) bool {
	get, err := store.db.Get(input.Key())
	input.Timestamp = time.Time{}
	gotThread, _ := newThreadFrom(get)
	gotThread.Timestamp = time.Time{}
	if err == nil && store.count == 1 && input.String() == gotThread.String() {
		return true
	}
	println("Count was incremented and/or DB does not contain thread")
	return false
}

func threadIsReadable(input Thread, returnValue uint64, store Store) bool {
	input.Post.No = 1
	input.Timestamp = time.Time{}
	get, err := store.db.Get(input.Key())
	outputAsThread, _ := newThreadFrom(get)
	outputAsThread.Timestamp = time.Time{}
	if err == nil && store.count == 2 && reflect.DeepEqual(input, outputAsThread) {
		return true
	}
	println("Count was not incremented and/or DB does not contain thread")
	return false
}

func thread() Thread {
	return NewThread(post(), "A subject")
}
