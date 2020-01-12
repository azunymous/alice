package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var tokenKey = []byte("KEYGOESHERE")

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
