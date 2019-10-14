package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alice-ws/alice/board"
	"github.com/alice-ws/alice/data"
	"github.com/alice-ws/alice/redisclient"
	"github.com/alice-ws/alice/users"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var userStore *users.Store
var threadStore *board.Store

type statusResponse struct {
	Status string `json:"status"`
}

type userResponse struct {
	Status   string `json:"status"`
	Username string `json:"username"`
	Error    string `json:"error"`
	Token    string `json:"token"`
}

type boardRequest struct {
	ThreadNo string     `json:"thread_no"`
	Post     board.Post `json:"post"`
	Type     string     `json:"type"`
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
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("jwt.key", "KEYGOESHERE")
	viper.SetDefault("users", map[string]string{"alice": "admin"})
	//TODO deal with below
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

func healthcheckHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	addHeaders(w)
	w.WriteHeader(http.StatusOK)
	status := statusResponse{
		Status: "HEALTHY",
	}

	_ = json.NewEncoder(w).Encode(status)
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

	post.Filename = header.Filename

	dir := viper.GetString("board.images.dir")
	tempImage, err := ioutil.TempFile(dir, "*-"+header.Filename)
	inMemoryImage, err := ioutil.ReadAll(image)
	if badRequest(err, w) {
		return
	}

	_, err = tempImage.Write(inMemoryImage)
	if badRequest(err, w) {
		return
	}

	post.Image = filepath.Base(tempImage.Name())

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

		post.Filename = header.Filename

		dir := viper.GetString("board.images.dir")
		tempImage, err := ioutil.TempFile(dir, "*-"+header.Filename)
		inMemoryImage, err := ioutil.ReadAll(image)
		if badRequest(err, w) {
			return
		}

		_, err = tempImage.Write(inMemoryImage)
		if badRequest(err, w) {
			return
		}

		post.Image = filepath.Base(tempImage.Name())
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
	router.GET("/healthcheck", healthcheckHandler)
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
	port := setup()
	log.Fatal(http.ListenAndServe(port, handler()))
}

func setup() string {
	configuration()
	port := viper.GetString("server.port")
	redisAddr := viper.GetString("redis.addr")
	tokenKey := []byte(viper.GetString("jwt.key"))
	rc, redisErr := redisclient.ConnectToRedis(redisAddr)
	if redisErr == nil {
		log.Printf("Connected to redis on " + redisAddr)
		userStore = users.NewStore(rc, tokenKey)
		threadStore = board.NewStore(rc)
	} else {
		log.Printf("Cannot connect to redis " + redisErr.Error() + " - falling back to in memory database")
		userStore = users.NewStore(data.NewMemoryDB(), tokenKey)
		threadStore = board.NewStore(data.NewMemoryDB())
	}
	log.Printf("Starting on " + port)
	return port
}
