package main

import (
	"encoding/json"
	"fmt"
	"github.com/alice-ws/alice/data"
	"github.com/alice-ws/alice/redisclient"
	"github.com/alice-ws/alice/users"
	"github.com/spf13/viper"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var userStore *users.Store

type statusResponse struct {
	Status string `json:"status"`
}

type userResponse struct {
	Status   string `json:"status"`
	Username string `json:"username"`
	Error    string `json:"error"`
	Token    string `json:"token"`
}

func configuration() {
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("jwt.key", "KEYGOESHERE")
	viper.SetDefault("users", map[string]string{"alice": "admin"})
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	viper.AddConfigPath("/alice/")        // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname") // call multiple times to add many search paths
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
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

	token, err := userStore.Register(email, username, password)

	w.Header().Set("Content-Type", "application/json")

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

	verified, err := userStore.Verify(token)

	if err != nil || !verified {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(userResponse{Status: "FAILURE"})
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
	router.POST("/verify", verifyUserHandler)

	return router
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
	} else {
		log.Printf("Cannot connect to redis " + redisErr.Error())
		userStore = users.NewStore(data.NewMemoryDB(), tokenKey)
	}
	log.Printf("Starting on " + port)
	return port
}
