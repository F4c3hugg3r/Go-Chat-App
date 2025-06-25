package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/server"
)

type Config struct {
	Port      int
	TimeLimit time.Duration
}

func main() {
	cfg := ParseFlags()

	service := server.NewChatService()
	plugin := server.RegisterPlugins(service)
	handler := server.NewServerHandler(service, plugin)

	router := handler.BuildMultiplexer()
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		for {
			time.Sleep(15 * time.Second)
			service.InactiveClientDeleter()
		}
	}()

	fmt.Println("Server running on port:", cfg.Port)
	log.Fatal(server.ListenAndServe())
}

func ParseFlags() Config {
	var cfg Config
	flag.IntVar(&cfg.Port, "port", 8080, "HTTP Server Port")
	flag.DurationVar(&cfg.TimeLimit, "timeLimit", 30, "Time limit for inactive clients in minutes")
	flag.Parse()
	return cfg
}
