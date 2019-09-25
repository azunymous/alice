package post

import (
	"github.com/alice-ws/alice/data"
	"reflect"
	"runtime"
	"testing"
)

func TestStore_AddThread(t *testing.T) {
	type fields struct {
		db    data.DB
		count uint64
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
			name: "first post is created with incremented post no and thread is added",
			fields: fields{
				db:    data.NewMemoryDB(),
				count: 0,
			},
			args: args{
				thread: thread(),
			},
			want:    countIsIncrementedForThread,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:    tt.fields.db,
				count: tt.fields.count,
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

func countIsIncrementedForThread(input Thread, returnValue uint64, store Store) bool {
	get, err := store.db.Get(input.Key())
	if err == nil && store.count == 1 && input.String() == get {
		return true
	}
	println("Count was not incremented and/or DB does not contain thread")
	return false
}

func thread() Thread {
	return NewThread(post(), "A subject")
}
