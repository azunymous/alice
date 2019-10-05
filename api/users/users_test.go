package users

import (
	"errors"
	"github.com/alice-ws/alice/data"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"testing"
)

func TestNewStore(t *testing.T) {
	type args struct {
		db  data.DB
		key []byte
	}
	tests := []struct {
		name string
		args args
		want *Store
	}{
		{
			name: "Simple user store with default in memory store",
			args: args{
				db:  nil,
				key: nil,
			},
			want: &Store{
				db:  data.NewMemoryDB(),
				key: nil,
			},
		},
		{
			name: "User store with given memory store and key",
			args: args{
				db:  data.NewMemoryDB(),
				key: []byte("GIVENKEY"),
			},
			want: &Store{
				db:  data.NewMemoryDB(),
				key: []byte("GIVENKEY"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStore(tt.args.db, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	type args struct {
		colonSeparatedUser string
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr bool
	}{
		{
			name: "Simple new user",
			args: args{
				colonSeparatedUser: "user:alice@alice.ws:alice:someHashedPassword",
			},
			want: &User{
				email:    "alice@alice.ws",
				Username: "alice",
				password: "someHashedPassword",
			},
			wantErr: false,
		},
		{
			name: "Fail to create user from invalid colon separated user",
			args: args{
				colonSeparatedUser: "user:Invalid:Number:Of:Colons:More:Than:Four",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.colonSeparatedUser)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Key(t *testing.T) {
	type fields struct {
		email    string
		Username string
		password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Username is the Key for a User",
			fields: fields{
				email:    "example@example.com",
				Username: "alice",
				password: "someHashedPassword",
			},
			want: "alice",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				email:    tt.fields.email,
				Username: tt.fields.Username,
				password: tt.fields.password,
			}
			if got := u.Key(); got != tt.want {
				t.Errorf("User.Key() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_String(t *testing.T) {
	type fields struct {
		email    string
		Username string
		password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Simple user is stringified correctly",
			fields: fields{
				email:    "alice@alice.ws",
				Username: "alice",
				password: "someHashedPassword",
			},
			want: "user:alice@alice.ws:alice:someHashedPassword",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				email:    tt.fields.email,
				Username: tt.fields.Username,
				password: tt.fields.password,
			}
			if got := u.String(); got != tt.want {
				t.Errorf("User.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

type failingDB struct{}

func (failingDB) Ping() bool {
	return false
}

func (failingDB) Set(data.KeyValue) error {
	return errors.New("cannot connect to DB")
}

func (failingDB) Get(string) (string, error) {
	return "", errors.New("cannot connect to DB")
}

func (failingDB) Remove(string) error {
	return errors.New("cannot connect to DB")
}

func TestStore_AnonymousRegister(t *testing.T) {
	type fields struct {
		db  data.DB
		key []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    func(Store, string, string) bool
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			want:    workingToken,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:  tt.fields.db,
				key: tt.fields.key,
			}

			username, got, err := store.AnonymousRegister()
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.AnonymousRegister() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if notW := !tt.want(*store, username, got); notW {
				t.Errorf("Store.AnonymousRegister() did not return a %v token; was %v", notW, got)
			}
		})
	}
}

func TestStore_Register(t *testing.T) {
	type fields struct {
		db  data.DB
		key []byte
	}
	type args struct {
		email    string
		username string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(Store, string, string) bool
		wantErr bool
	}{
		{
			name: "Simple user is registered",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				email:    "alice@alice.ws",
				username: "alice",
				password: "somePassword",
			},
			want:    workingToken,
			wantErr: false,
		},
		{
			name: "Too short username is not registered",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				email:    "alice@alice.ws",
				username: "",
				password: "somePassword",
			},
			want:    emptyTokenWithNoUserAdded,
			wantErr: true,
		},
		{
			name: "Too short email is not registered",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				email:    "a@b",
				username: "alice",
				password: "somePassword",
			},
			want:    emptyTokenWithNoUserAdded,
			wantErr: true,
		},
		{
			name: "Username with colons in is not registered",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				email:    "alice@alice.ws",
				username: "a:b:c",
				password: "somePassword",
			},
			want:    emptyTokenWithNoUserAdded,
			wantErr: true,
		},
		{
			name: "Email with colons in is not registered",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				email:    "a:b@example.com",
				username: "alice",
				password: "somePassword",
			},
			want:    emptyTokenWithNoUserAdded,
			wantErr: true,
		},
		{
			name: "Failing database causes user to not be registered and return an error",
			fields: fields{
				db:  failingDB{},
				key: []byte("aKey"),
			},
			args: args{
				email:    "alice@alice.ws",
				username: "alice",
				password: "somePassword",
			},
			want:    emptyTokenWithNoUserAdded,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:  tt.fields.db,
				key: tt.fields.key,
			}
			got, err := store.Register(tt.args.email, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if notW := !tt.want(*store, tt.args.username, got); notW {
				t.Errorf("Store.Register() did not return a %v token; was %v", notW, got)
			}
		})
	}
}

func workingToken(store Store, _, token string) bool {
	_, valid, err := store.Verify(token)
	if err != nil {
		println(err.Error())
	}
	return valid
}

func emptyToken(_ Store, _, token string) bool {
	if token == "" {
		return true
	}
	return false
}

func emptyTokenWithNoUserAdded(store Store, username, token string) bool {
	if _, err := store.db.Get(username); token == "" && err != nil {
		return true
	}
	return false
}

func TestStore_registerInDatabase(t *testing.T) {
	type fields struct {
		db  data.DB
		key []byte
	}
	type args struct {
		email    string
		username string
		pwd      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Simple user is registered in DB",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				email:    "alice@alice.ws",
				username: "alice",
				pwd:      "someHashedPassword",
			},
			wantErr: false,
		},
		{
			name: "User is not registered in failing DB",
			fields: fields{
				db:  failingDB{},
				key: []byte("aKey"),
			},
			args: args{
				email:    "alice@alice.ws",
				username: "alice",
				pwd:      "someHashedPassword",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:  tt.fields.db,
				key: tt.fields.key,
			}
			if err := store.registerInDatabase(tt.args.email, tt.args.username, tt.args.pwd); (err != nil) != tt.wantErr {
				t.Errorf("Store.registerInDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}

			u, getErr := store.db.Get(tt.args.username)
			if !tt.wantErr && getErr != nil {
				t.Errorf("Store.registerInDatabase() got user %v, with error  %v", u, getErr)
			}
		})
	}
}

func TestStore_getFromDatabase(t *testing.T) {
	type fields struct {
		db  data.DB
		key []byte
	}
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    User
		wantErr bool
	}{
		{
			name: "Get simple username",
			fields: fields{
				db:  mapStoreWithUser("alice", "alice@alice.ws", "someHashedPassword"),
				key: []byte("aKey"),
			},
			args: args{
				username: "alice",
			},
			want: User{
				email:    "alice@alice.ws",
				Username: "alice",
				password: "someHashedPassword",
			},
			wantErr: false,
		},
		{
			name: "Fail to get non-existent user",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				username: "rabbit",
			},
			want:    User{},
			wantErr: true,
		},
		{
			name: "Fail to get invalid user",
			fields: fields{
				db:  mapStoreWithUser("invalid", "invalid:user@example.com", ""),
				key: []byte("aKey"),
			},
			args: args{
				username: "invalid",
			},
			want:    User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:  tt.fields.db,
				key: tt.fields.key,
			}
			got, err := store.getFromDatabase(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.getFromDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.getFromDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mapStoreWithUser(username, email, password string) data.DB {
	db := data.NewMemoryDB()
	_ = db.Set(&User{Username: username, email: email, password: password})
	return db
}

func TestStore_Login(t *testing.T) {
	type fields struct {
		db  data.DB
		key []byte
	}
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(Store, string, string) bool
		wantErr bool
	}{
		{
			name: "Logins with existing user",
			fields: fields{
				db:  mapStoreWithUser("alice", "alice@alice.ws", hashPasswordIgnoringError("somePassword")),
				key: []byte("aKey"),
			},
			args: args{
				username: "alice",
				password: "somePassword",
			},
			want:    workingToken,
			wantErr: false,
		},
		{
			name: "Fails to login with non-existent user",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				username: "alice",
				password: "somePassword",
			},
			want:    emptyTokenWithNoUserAdded,
			wantErr: true,
		},
		{
			name: "Fails to login with empty password",
			fields: fields{
				db:  mapStoreWithUser("invalid", "a@b.com", hashPasswordIgnoringError("somePassword")),
				key: []byte("aKey"),
			},
			args: args{
				username: "invalid",
				password: "",
			},
			want:    emptyToken,
			wantErr: true,
		},
		{
			name: "Fails to login with incorrect password",
			fields: fields{
				db:  mapStoreWithUser("invalid", "a@b.com", hashPasswordIgnoringError("somePassword")),
				key: []byte("aKey"),
			},
			args: args{
				username: "invalid",
				password: "incorrect",
			},
			want:    emptyToken,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:  tt.fields.db,
				key: tt.fields.key,
			}
			got, err := store.Login(tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if notW := !tt.want(*store, tt.args.username, got); notW {
				t.Errorf("Store.Login() did not return a %v token; was %v", notW, got)
			}
		})
	}
}

func hashPasswordIgnoringError(password string) string {
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashedPwd)
}

func TestStore_Verify(t *testing.T) {
	type fields struct {
		db  data.DB
		key []byte
	}
	type args struct {
		userToken string
	}
	type want struct {
		username string
		valid    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Valid token is verified correctly",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				userToken: generateTokenIgnoringError("alice", "aKey"),
			},
			want:    want{"alice", true},
			wantErr: false,
		},
		{
			name: "Invalid token fails to verify",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				userToken: generateTokenIgnoringError("alice", "wrongKey"),
			},
			want:    want{"", false},
			wantErr: true,
		},
		{
			name: "Incorrectly formatted token fails to verify",
			fields: fields{
				db:  data.NewMemoryDB(),
				key: []byte("aKey"),
			},
			args: args{
				userToken: "only.twoparts",
			},
			want:    want{valid: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				db:  tt.fields.db,
				key: tt.fields.key,
			}
			gotUsername, got, err := store.Verify(tt.args.userToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want.valid || gotUsername != tt.want.username {
				t.Errorf("Store.Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO copy implementation here and use custom time argument for testing expiration as well
func generateTokenIgnoringError(username, key string) string {
	t, _ := generateToken(username, []byte(key))
	return t
}

func Test_generateToken(t *testing.T) {
	type args struct {
		username string
		key      []byte
	}
	tests := []struct {
		name         string
		args         args
		userStoreKey string
		want         func(Store, string, string) bool
		wantErr      bool
	}{
		{
			name: "Verifiable token is generated",
			args: args{
				username: "alice",
				key:      []byte("aKey"),
			},
			userStoreKey: "aKey",
			want:         workingToken,
			wantErr:      false,
		},
		{
			name: "Token generated is not verifiable if using wrong key",
			args: args{
				username: "alice",
				key:      []byte("WRONG_KEY"),
			},
			userStoreKey: "aKey",
			want:         invalidToken,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateToken(tt.args.username, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if notW := !tt.want(*NewStore(nil, []byte(tt.userStoreKey)), tt.args.username, got); notW {
				t.Errorf("Store.generateToken() did not return a %v token; was %v", notW, got)
			}
		})
	}
}
func invalidToken(store Store, _, token string) bool {
	_, valid, err := store.Verify(token)
	if err != nil {
		println(err.Error())
	}
	return !valid
}
