package main

import (
	"encoding/json"
	"fmt"
	"github.com/alice-ws/alice/board"
	"github.com/alice-ws/alice/data"
	"github.com/alice-ws/alice/dependencies"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var dependencyManagement dependencies.Dependencies

var threadStore *board.Store
var mediaRepo data.MediaRepo

type statusResponse struct {
	Status string `json:"status"`
}

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
	viper.SetDefault("board.ID", "/obj/")
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

func handler() http.Handler {
	router := httprouter.New()
	router.GET("/", homePageHandler)
	router.GET("/ready", readyHandler)
	router.GET("/thread/all", getAllThreadsHandler)
	router.POST("/thread", addThreadHandler)
	router.GET("/thread", getThreadHandler)
	router.POST("/post", addPostHandler)

	return cors.Default().Handler(router)
}

func addHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
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
	mc := dependencyManagement.GetImageRepository()
	mediaRepo = mc

	db := dependencyManagement.GetDB()
	boardID := viper.GetString("board.ID")
	threadStore = board.NewStore(boardID, db, db)

	log.Printf("Starting on " + port)
	return port
}
