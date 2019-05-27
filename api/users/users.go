package users

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"
)

type DB struct {
	store UserStore
	key   []byte
}

func NewDB(store UserStore, key []byte) *DB {
	if store == nil {
		store = NewMemoryStore()
	}

	return &DB{
		store: store,
		key:   key,
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

func NewUser(colonSeparatedUser string) (*User, error) {
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

func (u *User) String() string {
	return fmt.Sprintf("user:%s:%s:%s", u.email, u.Username, u.password)
}

type UserStore interface {
	Ping() bool
	Add(User) error
	Get(username string) (User, error)
	Remove(User) error
}

func (DB *DB) Register(email, username, password string) (string, error) {
	if len(email) < 1 || len(username) < 1 || len(password) < 1 {
		return "", errors.New("fields too short")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	DB.registerInDatabase(email, username, string(hashedPwd))
	token, tokenErr := generateToken(username, DB.key)
	if tokenErr != nil {
		return "", tokenErr
	}
	return token, nil
}

func (DB *DB) registerInDatabase(email, username, pwd string) {
	err := DB.store.Add(User{
		email:    email,
		Username: username,
		password: pwd,
	})
	if err != nil {
		log.Printf("Error adding user %s (%s): %v", username, email, err)
	}
}

func (DB *DB) Login(username, password string) (string, error) {
	user, getErr := DB.store.Get(username)
	if getErr != nil {
		return "", errors.New("error logging in: " + getErr.Error())
	}
	pwdErr := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if pwdErr != nil {
		return "", pwdErr
	}
	tokenString, err := generateToken(username, DB.key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (DB *DB) Verify(userToken string) (bool, error) {
	claims := &Claim{}
	parsedTkn, err := jwt.ParseWithClaims(userToken, claims, func(t *jwt.Token) (i interface{}, e error) {
		return DB.key, nil
	})
	if err != nil {
		return false, err
	}
	if !parsedTkn.Valid {
		return false, errors.New("token not valid")
	}
	println(claims.Username, " verified token")
	return true, nil
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
