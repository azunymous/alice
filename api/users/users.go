package users

import (
	"errors"
	"fmt"
	"github.com/alice-ws/alice/anon"
	"github.com/alice-ws/alice/data"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"
)

type Store struct {
	db  data.DB
	key []byte
}

func NewStore(db data.DB, key []byte) *Store {
	if db == nil {
		db = data.NewMemoryDB()
	}

	return &Store{
		db:  db,
		key: key,
	}
}

type User struct {
	email    string
	Username string
	password string
}

type Claim struct {
	Username string
	jwt.StandardClaims
}

func New(colonSeparatedUser string) (*User, error) {
	split := strings.Split(colonSeparatedUser, ":")
	if len(split) != 4 {
		return nil, errors.New("user not registered correctly")
	}
	return &User{
		email:    split[1],
		Username: split[2],
		password: split[3],
	}, nil
}

func (u *User) Key() string {
	return u.Username
}

func (u *User) String() string {
	return fmt.Sprintf("user:%s:%s:%s", u.email, u.Username, u.password)
}

func (store *Store) AnonymousRegister() (string, string, error) {
	email, password := anon.Defaults()

	username := anon.GenerateUsername(time.Now().UnixNano())
	token, err := store.Register(email, username, password)
	return username, token, err
}

func (store *Store) Register(email, username, password string) (string, error) {
	if valid, err := fieldsAreValid(email, username, password); !valid {
		return "", err
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	err = store.registerInDatabase(email, username, string(hashedPwd))
	if err != nil {
		return "", err
	}
	token, tokenErr := generateToken(username, store.key)
	if tokenErr != nil {
		return "", tokenErr
	}
	return token, nil
}

func fieldsAreValid(email, username, password string) (bool, error) {
	if len(email) < 5 || len(username) < 3 || len(password) < 1 {
		return false, errors.New("fields too short")
	}
	if strings.ContainsRune(email, ':') || strings.ContainsRune(username, ':') {
		return false, errors.New("email or username contain invalid character: ':'")
	}
	return true, nil
}

func (store *Store) registerInDatabase(email, username, pwd string) error {
	err := store.db.Add(&User{
		email:    email,
		Username: username,
		password: pwd,
	})
	if err != nil {
		log.Printf("Error adding user %s (%s): %v", username, email, err)
		return err
	}
	return nil
}

func (store *Store) getFromDatabase(username string) (User, error) {
	result, err := store.db.Get(username)
	if err != nil {
		return User{}, errors.New("error getting user")
	}
	user, newUserErr := New(result)
	if newUserErr != nil {
		return User{}, errors.New("error reading stored user " + newUserErr.Error())
	}
	return *user, nil
}

func (store *Store) Login(username, password string) (string, error) {
	user, getErr := store.getFromDatabase(username)
	if getErr != nil {
		return "", errors.New("error logging in: " + getErr.Error())
	}
	pwdErr := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if pwdErr != nil {
		return "", pwdErr
	}
	tokenString, err := generateToken(username, store.key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (store *Store) Verify(userToken string) (string, bool, error) {
	claims := &Claim{}
	parsedTkn, err := jwt.ParseWithClaims(userToken, claims, func(t *jwt.Token) (i interface{}, e error) {
		return store.key, nil
	})
	if err != nil {
		return "", false, err
	}
	if !parsedTkn.Valid {
		return "", false, errors.New("token not valid")
	}
	println(claims.Username, " verified token")
	return claims.Username, true, nil
}

func generateToken(username string, key []byte) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	// Create the JWT claims, which includes the username and expiry time
	claim := Claim{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	// Create the JWT string
	tokenString, err := token.SignedString(key)
	return tokenString, err
}
