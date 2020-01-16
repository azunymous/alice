package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

type Board struct {
	Host   string `json:"host"`
	Images string `json:"images"`
}

func homePageHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	addHeaders(w)
	_, _ = fmt.Fprintf(w, `{"V" : "1", "data" : "ALICE OVERBOARD API"}`)
}

func overboardHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	addHeaders(w)
	var boards map[string]Board
	err := viper.UnmarshalKey("boards", &boards)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, `{"ERROR" : "Could not read configuration"}`)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(boards)
}

func handler() http.Handler {
	router := httprouter.New()
	router.GET("/", homePageHandler)
	router.GET("/boards", overboardHandler)
	return cors.Default().Handler(router)
}

func addHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func main() {
	viper.SetDefault("server.port", ":9090")

	viper.SetDefault("boards", map[string]Board{
		"/obj/": {Host: "http://localhost:8080", Images: "/images/"},
	})

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.AddConfigPath("/config/")
	viper.AddConfigPath("/overboard/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.alice") // call multiple times to add many search paths
	err := viper.ReadInConfig()         // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		log.Printf("Config file error: %s \n", err)
	} else {
		log.Printf("Watching config file")
		viper.WatchConfig()
	}

	port := viper.GetString("server.port")
	log.Fatal(http.ListenAndServe(port, handler()))
}
