package main

import (
	"encoding/json"
	"errors"
	"github.com/alice-ws/alice/board"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

type userResponse struct {
	Status   string `json:"status"`
	Username string `json:"username"`
	Error    string `json:"error"`
	Token    string `json:"token"`
}

type boardResponse struct {
	Status string       `json:"status"`
	No     string       `json:"no"`
	Thread board.Thread `json:"thread"`
	Type   string       `json:"type"`
}

const (
	Thread = "THREAD"
	Post   = "POST"
)

func getAllThreadsHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	t, err := threadStore.GetAllThreads()

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(boardResponse{Status: "FAILURE"})
		return
	}

	addHeaders(w)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(t)

}

func addThreadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := r.ParseMultipartForm(10 << 20)

	if badRequest(err, w) {
		return
	}
	name := r.FormValue("name")
	email := r.FormValue("email")
	subject := r.FormValue("subject")
	comment := r.FormValue("comment")
	post := board.CreatePost(name, email, comment)

	_, header, err := r.FormFile("image")
	if badRequest(err, w) {
		return
	}
	image, err := header.Open()
	if badRequest(err, w) {
		return
	}

	URI, err := mediaRepo.Store(image, dependencyManagement.ImageGroup(), mediaRepo.GenerateUniqueName(header.Filename), header.Size)

	post.Filename = header.Filename

	if badRequest(err, w) {
		return
	}

	post.Image = URI

	log.Printf("Add Thread: %v with subject %s", post, subject)
	t := board.NewThread(post, subject)

	_, err = threadStore.AddThread(t)

	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE"})
		return
	}

	addHeaders(w)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(boardResponse{
		Status: "SUCCESS",
	})
}

func badRequest(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Printf("Error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(boardResponse{Status: "FAILURE"})
		return true
	}
	return false
}

func getThreadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	threadNo := r.URL.Query().Get("no")
	if threadNo == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(boardResponse{Status: "FAILURE"})
		return
	}
	t, err := threadStore.GetThread(threadNo)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(boardResponse{Status: "FAILURE"})
		return
	}

	addHeaders(w)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(boardResponse{Status: "SUCCESS", No: threadNo, Thread: t, Type: Thread})

}

func addPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := r.ParseMultipartForm(10 << 20)

	if badRequest(err, w) {
		return
	}
	name := r.FormValue("name")
	email := r.FormValue("email")
	thread := r.FormValue("threadNo")
	comment := r.FormValue("comment")

	log.Printf("Creating post with fields %s, %s, %s in thread %s", name, email, comment, thread)
	post := board.CreatePost(name, email, comment)

	_, header, err := r.FormFile("image")

	if err == nil {
		image, err := header.Open()
		if badRequest(err, w) {
			return
		}

		URI, err := mediaRepo.Store(image, dependencyManagement.ImageGroup(), mediaRepo.GenerateUniqueName(header.Filename), header.Size)

		post.Filename = header.Filename

		if badRequest(err, w) {
			return
		}

		post.Image = URI
	}

	if !post.IsValid() {
		badRequest(errors.New("Invalid Post: "+post.String()), w)
		return
	}

	log.Printf("Added Post: %v in thread %s", post, thread)
	_, err = threadStore.AddPost(thread, post)

	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE", Error: err.Error()})
		return
	}

	addHeaders(w)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(boardResponse{
		Status: "SUCCESS",
	})
}
