package main

import (
	"encoding/json"
	"fmt"
	"github.com/alice-ws/alice/redisclient"
	"github.com/alice-ws/alice/users"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var us users.UserStore = nil
var userDB = users.NewDB(us)

type statusResponse struct {
	Status string `json:"status"`
}

type userResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func homePageHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"V" : "1", "data" : "ALICE API"}`)
}

func healthcheckHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	status := statusResponse{
		Status: "HEALTHY",
	}

	_ = json.NewEncoder(w).Encode(status)
}

func registerHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_ = r.ParseForm()
	var (
		email    = r.Form.Get("email")
		username = r.Form.Get("username")
		password = r.Form.Get("password")
	)

	token, err := userDB.Register(email, username, password)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE: " + err.Error()})
	}
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(userResponse{
		Status: "SUCCESS",
		Token:  token,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_ = r.ParseForm()
	var (
		username = r.Form.Get("username")
		password = r.Form.Get("password")
	)

	token, err := userDB.Login(username, password)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userResponse{
		Status: "SUCCESS",
		Token:  token,
	})

}

func handler() http.Handler {
	router := httprouter.New()
	router.GET("/", homePageHandler)
	router.GET("/healthcheck", healthcheckHandler)
	router.POST("/register", registerHandler)
	router.POST("/login", loginHandler)

	return router
}

func main() {
	rc, redisErr := redisclient.ConnectToRedis("localhost:6379")
	if redisErr == nil {
		us = rc
	} else {
		log.Printf("Cannot connect to redis " + redisErr.Error())
	}

	log.Printf("Starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler()))
}
