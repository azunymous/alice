package main

import (
	"encoding/json"
	"github.com/alice-ws/alice/board"
	"github.com/alice-ws/alice/data"
	"github.com/alice-ws/alice/users"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

var tokenKey = []byte("KEYGOESHERE")

func Test_healthcheckHandler(t *testing.T) {
	endpoint := "/healthcheck"
	method := "GET"

	rr := createRequestAndServe(method, endpoint, nil, requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	// Check the response body is what we expect.
	expected := `{"status" :"HEALTHY"}`
	checkBody(rr.Body.String(), expected, t)
}

func Test_homepageHandler(t *testing.T) {
	endpoint := "/"
	method := "GET"

	rr := createRequestAndServe(method, endpoint, nil, requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	// Check the response body is what we expect.
	expected := `{"V" : "1", "data" : "ALICE API"}`
	checkBody(rr.Body.String(), expected, t)
}

func Test_AnonRegisterSuccess(t *testing.T) {
	userStore = users.NewStore(nil, tokenKey)
	endpoint := "/anonregister"
	method := "POST"

	d := url.Values{}

	rr := createRequestAndServe(method, endpoint, strings.NewReader(d.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusCreated, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)
	verifyToken(response, t)

	_, err := userStore.Login(response.Username, "password")

	if err != nil {
		t.Errorf("Failed to login with registered user: %v", err)
	}
}

func Test_RegisterSuccess(t *testing.T) {
	userStore = users.NewStore(nil, tokenKey)
	endpoint := "/register"
	method := "POST"

	values := url.Values{}
	values.Set("email", "alice@example.com")
	values.Set("username", "alice")
	values.Set("password", "Password123")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(values.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusCreated, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)
	verifyToken(response, t)

	_, err := userStore.Login("alice", "Password123")

	if err != nil {
		t.Errorf("Failed to login with registered user: %v", err)
	}
}

func Test_RegisterFailureMissingField(t *testing.T) {
	userStore = users.NewStore(nil, tokenKey)
	endpoint := "/register"
	method := "POST"

	values := url.Values{}
	values.Set("email", "alice@example.com")
	values.Set("password", "Password123")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(values.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusBadRequest, t)

	_, err := userStore.Login("alice", "Password123")

	if err == nil {
		t.Errorf("Succeded to login with registered user, should've failed")
	}
}

func Test_RegisterFailureFieldsEmpty(t *testing.T) {
	userStore = users.NewStore(nil, tokenKey)
	endpoint := "/register"
	method := "POST"

	values := url.Values{}
	values.Set("email", "")
	values.Set("username", "alice")
	values.Set("password", "")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(values.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusBadRequest, t)

	_, err := userStore.Login("alice", "Password123")

	if err == nil {
		t.Errorf("Succeded to login with registered user, should've failed")
	}
}

func Test_LoginSuccess(t *testing.T) {
	store := data.NewMemoryDB()
	password, _ := bcrypt.GenerateFromPassword([]byte("Password123"), 10)
	user, _ := users.New("user:alice@example.com:alice:" + string(password))
	_ = store.Set(user)
	userStore = users.NewStore(store, tokenKey)

	endpoint := "/login"
	method := "POST"

	values := url.Values{}
	values.Set("username", "alice")
	values.Set("password", "Password123")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(values.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)

	verifyToken(response, t)
}

func Test_LoginFailure(t *testing.T) {
	store := data.NewMemoryDB()
	password, _ := bcrypt.GenerateFromPassword([]byte("Password123"), 10)
	user, _ := users.New("user:alice@example.com:alice:" + string(password))
	_ = store.Set(user)
	userStore = users.NewStore(store, tokenKey)

	endpoint := "/login"
	method := "POST"

	values := url.Values{}
	values.Set("username", "alice")
	values.Set("password", "IncorrectPassword")

	rr := createRequestAndServe(method, endpoint, strings.NewReader(values.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusUnauthorized, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "FAILURE", t)
	checkError(response, "crypto/bcrypt: hashedPassword is not the hash of the given password", t)

}

func Test_FullPathToVerifyHandlerSuccess(t *testing.T) {
	userStore = users.NewStore(nil, tokenKey)
	endpoint := "/verify"
	method := "POST"

	registerToken, _ := userStore.Register("alice@alice.ws", "alice", "Password123")
	loginToken, _ := userStore.Login("alice", "Password123")

	registerData := url.Values{}
	registerData.Set("token", registerToken)

	loginData := url.Values{}
	loginData.Set("token", loginToken)

	registerR := createRequestAndServe(method, endpoint, strings.NewReader(registerData.Encode()), requestCreatorForm)
	loginR := createRequestAndServe(method, endpoint, strings.NewReader(loginData.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(registerR.Code, http.StatusOK, t)
	checkStatusCode(loginR.Code, http.StatusOK, t)

	registerResponse := &userResponse{}
	_ = json.Unmarshal(registerR.Body.Bytes(), registerResponse)
	loginResponse := &userResponse{}
	_ = json.Unmarshal(loginR.Body.Bytes(), loginResponse)
	const want = "SUCCESS"
	if registerResponse.Status != want {
		t.Errorf("Verify response returned wrong status: got %v want %s", registerResponse.Status, want)
	}
	if loginResponse.Status != want {
		t.Errorf("Verify response returned wrong status: got %v want %s", loginResponse.Status, want)
	}

}

func Test_InvalidVerifyHandlerFailure(t *testing.T) {
	userStore = users.NewStore(nil, tokenKey)
	endpoint := "/verify"
	method := "POST"

	registerToken, _ := userStore.Register("alice@alice.ws", "alice", "Password123")
	split := strings.Split(registerToken, ".")
	registerToken = split[0] + "." + split[2] + "." + split[1]
	registerData := url.Values{}
	registerData.Set("token", registerToken)

	registerR := createRequestAndServe(method, endpoint, strings.NewReader(registerData.Encode()), requestCreatorForm)

	// Check the status code is what we expect.
	checkStatusCode(registerR.Code, http.StatusUnauthorized, t)

	registerResponse := &userResponse{}
	_ = json.Unmarshal(registerR.Body.Bytes(), registerResponse)
	const want = "FAILURE"
	if registerResponse.Status != want {
		t.Errorf("Verify response returned wrong status: got %v want %s", registerResponse.Status, want)
	}

}

func Test_AddThreadSuccess(t *testing.T) {
	threadStore = board.NewStore(nil)
	endpoint := "/thread"
	method := "POST"
	thread := board.NewThread(examplePost(), "a subject")
	rr := createRequestAndServe(method, endpoint, strings.NewReader(thread.String()), requestCreatorJson)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusCreated, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)
	threadInDB, err := threadStore.GetThread(strconv.FormatUint(thread.No, 10))
	if err != nil || !reflect.DeepEqual(threadInDB, thread) {
		t.Errorf("Thread in DB incorrect: got %v want %s", threadInDB, thread)
	}
}

func Test_GetThreadSuccess(t *testing.T) {
	threadStore = board.NewStore(nil)
	threadInDB := board.NewThread(examplePost(), "a subject")
	_, _ = threadStore.AddThread(threadInDB) // Thread number is 0
	endpoint := "/thread"
	method := "GET"
	endpoint = addQueryParam(endpoint, "no", "0")
	rr := createRequestAndServe(method, endpoint, nil, requestCreatorJson)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusOK, t)

	response := &boardResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)

	if !reflect.DeepEqual(response.Thread, threadInDB) || response.Type != "THREAD" {
		t.Errorf("Thread in response incorrect: got %v want %s", response.Thread, threadInDB)
		t.Logf("Full response: %v", response)
	}
}

//noinspection ALL
func Test_AddPostSuccess(t *testing.T) {
	threadStore = board.NewStore(nil)
	threadInDB := board.NewThread(examplePost(), "a subject")
	_, _ = threadStore.AddThread(threadInDB) // Thread number is 0

	endpoint := "/post"
	method := "POST"

	post := examplePost()
	post.No = 1
	request := boardRequest{
		ThreadNo: "0",
		Post:     post,
		Type:     "POST",
	}

	bytes, _ := json.Marshal(request)

	rr := createRequestAndServe(method, endpoint, strings.NewReader(string(bytes)), requestCreatorJson)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusCreated, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)
	checkResponse(response, "SUCCESS", t)
	threadInDB, err := threadStore.GetThread(request.ThreadNo)
	if err != nil || len(threadInDB.Replies) != 1 || !reflect.DeepEqual(threadInDB.Replies[0], post) {
		t.Errorf("Thread in DB incorrect: got %v want %s", threadInDB.Replies[0], post)
	}
}

//noinspection ALL
func Test_AddPostSuccess_generatesPostNo(t *testing.T) {
	threadStore = board.NewStore(nil)
	threadInDB := board.NewThread(examplePost(), "a subject")
	_, _ = threadStore.AddThread(threadInDB) // Thread number is 0

	endpoint := "/post"
	method := "POST"

	post := examplePost()
	post.No = 99
	request := boardRequest{
		ThreadNo: "0",
		Post:     post,
		Type:     "POST",
	}

	bytes, _ := json.Marshal(request)

	rr := createRequestAndServe(method, endpoint, strings.NewReader(string(bytes)), requestCreatorJson)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusCreated, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)

	wantPost := examplePost()
	wantPost.No = 1
	checkResponse(response, "SUCCESS", t)
	threadInDB, err := threadStore.GetThread(string(request.ThreadNo))
	if err != nil || len(threadInDB.Replies) != 1 || !reflect.DeepEqual(threadInDB.Replies[0], wantPost) {
		t.Errorf("Thread in DB incorrect: got %v want %s", threadInDB.Replies[0], wantPost)
	}
}

func Test_AddPostFailure_InvalidThreadNo(t *testing.T) {
	threadStore = board.NewStore(nil)
	threadInDB := board.NewThread(examplePost(), "a subject")
	_, _ = threadStore.AddThread(threadInDB) // Thread number is 0

	endpoint := "/post"
	method := "POST"

	post := examplePost()
	request := boardRequest{
		ThreadNo: "99",
		Post:     post,
		Type:     "POST",
	}

	bytes, _ := json.Marshal(request)

	rr := createRequestAndServe(method, endpoint, strings.NewReader(string(bytes)), requestCreatorJson)

	// Check the status code is what we expect.
	checkStatusCode(rr.Code, http.StatusBadRequest, t)

	response := &userResponse{}
	_ = json.Unmarshal(rr.Body.Bytes(), response)

	checkResponse(response, "FAILURE", t)
	threadInDB, err := threadStore.GetThread("0")
	if err != nil || len(threadInDB.Replies) != 0 {
		t.Errorf("Thread in DB incorrect: got %v want no replies", threadInDB)
	}
}

// Test Utilities
var h = handler()

func createRequestAndServe(method string, hitEndpoint string, params io.Reader, requestCreator func(string, string, io.Reader) *http.Request) *httptest.ResponseRecorder {
	req := requestCreator(method, hitEndpoint, params)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func requestCreatorForm(method, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if body != nil && method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	return req
}

func requestCreatorJson(method, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if body != nil && method == "POST" {
		req.Header.Add("Content-Type", "application/json")
	}
	return req
}

func addQueryParam(endpoint, key, value string) string {
	parsedEndpoint, _ := url.Parse(endpoint)
	q := parsedEndpoint.Query()
	q.Add(key, value)
	log.Printf("endpoint with query %s", parsedEndpoint.String())
	parsedEndpoint.RawQuery = q.Encode()
	return parsedEndpoint.String()
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

func checkError(response *userResponse, expectedError string, t *testing.T) {
	if response.Error != expectedError {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response.Error, expectedError)
	}
}

func verifyToken(response *userResponse, t *testing.T) {
	_, b, e := userStore.Verify(response.Token)
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

func examplePost() board.Post {
	return board.NewPost(0, time.Unix(0, 0), "Anonymous", "", "Hello World!", "/path/0", "file.png", "")
}
