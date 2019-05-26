package main

import (
	"encoding/json"
	"github.com/alice-ws/alice/users"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Test_healthcheckHandler(t *testing.T) {
	endpoint := "/healthcheck"
	method := "GET"

	rr := createRequestAndServe(method, endpoint, nil)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	// Check the response body is what we expect.
	expected := `{"status" :"HEALTHY"}`
	checkBody(rr.Body.String(), expected, t)
}

func Test_homepageHandler(t *testing.T) {
	endpoint := "/"
	method := "GET"

	rr := createRequestAndServe(method, endpoint, nil)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	// Check the response body is what we expect.
	expected := `{"V" : "1", "data" : "ALICE API"}`
	checkBody(rr.Body.String(), expected, t)
}

func Test_RegisterSuccess(t *testing.T) {
	userDB = users.NewDB(nil)
	endpoint := "/register"
	method := "POST"

	data := url.Values{}
	data.Set("email", "alice@example.com")
	data.Set("username", "alice")
	data.Set("password", "Password123")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(data.Encode()))

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusCreated, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)
	verifyToken(response, t)

	_, err := userDB.Login("alice", "Password123")

	if err != nil {
		t.Errorf("Failed to login with registered user: %v", err)
	}
}

func Test_RegisterFailureMissingField(t *testing.T) {
	userDB = users.NewDB(nil)
	endpoint := "/register"
	method := "POST"

	data := url.Values{}
	data.Set("email", "alice@example.com")
	data.Set("password", "Password123")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(data.Encode()))

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusBadRequest, t)

	_, err := userDB.Login("alice", "Password123")

	if err == nil {
		t.Errorf("Succeded to login with registered user, should've failed")
	}
}

func Test_LoginSuccess(t *testing.T) {
	store := users.NewMemoryStore()
	password, _ := bcrypt.GenerateFromPassword([]byte("Password123"), 10)
	user, _ := users.NewUser("user:alice@example.com:alice:" + string(password))
	_ = store.Add(*user)
	userDB = users.NewDB(store)

	endpoint := "/login"
	method := "POST"

	data := url.Values{}
	data.Set("username", "alice")
	data.Set("password", "Password123")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(data.Encode()))

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)

	verifyToken(response, t)
}

func Test_LoginFailure(t *testing.T) {
	store := users.NewMemoryStore()
	password, _ := bcrypt.GenerateFromPassword([]byte("Password123"), 10)
	user, _ := users.NewUser("user:alice@example.com:alice:" + string(password))
	_ = store.Add(*user)
	userDB = users.NewDB(store)

	endpoint := "/login"
	method := "POST"

	data := url.Values{}
	data.Set("username", "alice")
	data.Set("password", "IncorrectPassword")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(data.Encode()))

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusUnauthorized, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "FAILURE: crypto/bcrypt: hashedPassword is not the hash of the given password", t)

}

// Test Utilities
var h = handler()

func createRequestAndServe(method string, hitEndpoint string, params io.Reader) *httptest.ResponseRecorder {
	req := getRequest(method, hitEndpoint, params)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func getRequest(method, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if body != nil && method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	return req
}

func checkStatusCode(got, expected int, t *testing.T) {
	// Check the status code is what we expect.
	if got != expected {
		t.Errorf("handler returned wrong status code: got %v want %v",
			got, expected)
	}
}

func checkBody(got, expected string, t *testing.T) {
	if equal, _ := JSONBytesEqual(got, expected); !equal {
		t.Errorf("handler returned unexpected body: got %v want %v",
			got, expected)
	}
}

func checkResponse(response *userResponse, expectedStatus string, t *testing.T) {
	if response.Status != expectedStatus {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response.Status, expectedStatus)
	}
}

func verifyToken(response *userResponse, t *testing.T) {
	b, e := userDB.Verify(response.Token)
	if !b {
		t.Errorf("token was invalid: got %v - error %v",
			response.Token, e)
	}
}

func JSONBytesEqual(s1, s2 string) (bool, error) {
	a, b := []byte(s1), []byte(s2)

	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}
