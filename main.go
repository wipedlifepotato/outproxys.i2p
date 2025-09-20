package main

import (
	"github.com/spf13/viper"
	"log"
	"net/http"
	"outproxys/config"
	"outproxys/goroutines"
	"outproxys/handlers"
	"strconv"
	"time"
)

func main() {
	config.LoadConfig()
	log.Println("Run server")

	http.HandleFunc("/", handlers.MainHandle)

	http.HandleFunc("/add", handlers.AddHandle)

	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))

	go goroutines.MonitorProxies()
	host := viper.GetString("server.host")
	port := strconv.Itoa(viper.GetInt("server.port"))

	s := &http.Server{
		Addr:           host + ":" + port,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Listen on host " + host)
	log.Println("Listen server on port " + port)
	log.Fatal(s.ListenAndServe())
}
