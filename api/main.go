package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alice-ws/alice/board"
	"github.com/alice-ws/alice/data"
	"github.com/alice-ws/alice/dependencies"
	"github.com/alice-ws/alice/users"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var dependencyManagement dependencies.Dependencies

var userStore *users.Store
var threadStore *board.Store
var mediaRepo data.MediaRepo

type statusResponse struct {
	Status string `json:"status"`
}

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

func configuration() {
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.fallbacks.markUnhealthy", false)
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.timeout", "2s")
	viper.SetDefault("minio.addr", "localhost:9000")
	viper.SetDefault("minio.timeout", "2s")
	_ = viper.BindEnv("minio.access", "MINIO_ACCESS_KEY")
	_ = viper.BindEnv("minio.secret", "MINIO_SECRET_KEY")
	viper.SetDefault("minio.access", "minio")
	viper.SetDefault("minio.secret", "insecure")
	viper.SetDefault("jwt.key", "KEYGOESHERE")
	viper.SetDefault("users", map[string]string{"alice": "admin"})

	dir, _ := os.Getwd()
	viper.SetDefault("board.images.dir", filepath.Join(filepath.Dir(dir), "/web/public/images"))

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.AddConfigPath("/config/")
	viper.AddConfigPath("/alice/")      // path to look for the config file in
	viper.AddConfigPath("$HOME/.alice") // call multiple times to add many search paths
	err := viper.ReadInConfig()         // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		log.Printf("Config file error: %s \n", err)
	} else {
		log.Printf("Watching config file")
		viper.WatchConfig()
	}
}

func homePageHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	addHeaders(w)
	_, _ = fmt.Fprintf(w, `{"V" : "1", "data" : "ALICE API"}`)
}

func liveHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	addHeaders(w)
	w.WriteHeader(http.StatusOK)
	status := statusResponse{
		Status: "HEALTHY",
	}

	_ = json.NewEncoder(w).Encode(status)
}

func readyHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	dependencyManagement.MarkFallbacksUnhealthy(viper.GetBool("server.fallbacks.markUnhealthy"))
	dependenciesList := dependencyManagement.String()

	addHeaders(w)
	if dependencyManagement.Healthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, _ = fmt.Fprintf(w, dependenciesList)
}

func anonRegisterHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	username, token, err := userStore.AnonymousRegister()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(userResponse{
		Status:   "SUCCESS",
		Username: username,
		Token:    token,
	})
}

func registerHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_ = r.ParseForm()
	var (
		email    = r.Form.Get("email")
		username = r.Form.Get("username")
		password = r.Form.Get("password")
	)

	token, err := userStore.Register(email, username, password)

	addHeaders(w)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE", Error: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(userResponse{
		Status:   "SUCCESS",
		Username: username,
		Token:    token,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_ = r.ParseForm()
	var (
		username = r.Form.Get("username")
		password = r.Form.Get("password")
	)

	token, err := userStore.Login(username, password)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE", Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userResponse{
		Status:   "SUCCESS",
		Username: username,
		Token:    token,
	})

}

func verifyUserHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_ = r.ParseForm()
	var (
		token = r.Form.Get("token")
	)

	username, verified, err := userStore.Verify(token)

	if err != nil || !verified {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE"})
		return
	}

	addHeaders(w)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(userResponse{
		Status:   "SUCCESS",
		Username: username,
		Token:    token,
	})

}

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
	_ = json.NewEncoder(w).Encode(userResponse{
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
	_ = json.NewEncoder(w).Encode(boardResponse{No: threadNo, Thread: t, Type: Thread})

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

	log.Printf("Add Post: %v in thread %s", post, thread)
	_, err = threadStore.AddPost(thread, post)

	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE", Error: err.Error()})
		return
	}

	addHeaders(w)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(userResponse{
		Status: "SUCCESS",
	})
}

func addHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func handler() http.Handler {
	router := httprouter.New()
	router.GET("/", homePageHandler)
	router.GET("/ready", readyHandler)
	router.POST("/register", registerHandler)
	router.POST("/anonregister", anonRegisterHandler)
	router.POST("/login", loginHandler)
	router.POST("/verify", verifyUserHandler)
	router.GET("/thread/all", getAllThreadsHandler)
	router.POST("/thread", addThreadHandler)
	router.GET("/thread", getThreadHandler)
	router.POST("/post", addPostHandler)

	return cors.Default().Handler(router)
}

func main() {
	go serveLiveness()
	dependencyManagement = dependencies.Setup()
	port := setup()
	log.Fatal(http.ListenAndServe(port, handler()))
}

func serveLiveness() {
	router := httprouter.New()
	router.GET("/live", liveHandler)
	log.Fatal(http.ListenAndServe(":8081", router))
}

func setup() string {
	configuration()
	port := viper.GetString("server.port")
	tokenKey := []byte(viper.GetString("jwt.key"))

	mc := dependencyManagement.GetImageRepository()
	mediaRepo = mc

	db := dependencyManagement.GetDB()

	userStore = users.NewStore(db, tokenKey)
	threadStore = board.NewStore("/obj/", db, db)

	log.Printf("Starting on " + port)
	return port
}
