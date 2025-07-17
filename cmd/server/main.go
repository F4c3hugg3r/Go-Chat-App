package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/api"
	chat "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/chat"
)

type Config struct {
	Port      int
	TimeLimit time.Duration
	maxUsers  int
}

func main() {
	cfg := ParseFlags()
	service := chat.NewChatService(cfg.maxUsers)
	plugin := chat.RegisterPlugins(service)
	handler := api.NewServerHandler(service, plugin)
	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	setUp(server, handler, cfg.TimeLimit, wg, ctx)

	interChan := make(chan os.Signal, 2)
	signal.Notify(interChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	wg.Add(1)

	go interruptListener(interChan, server, wg, cancel)

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer ln.Close()
	log.Println("Server running on port:", cfg.Port)

	err = server.Serve(ln)
	if err != nil {
		log.Println(err.Error())
	}

	wg.Wait()
}

// interruptListener sends a cancel() signal and shuts down the server gracefully if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, server *http.Server, wg *sync.WaitGroup, cancel context.CancelFunc) {
	defer wg.Done()

	<-interChan

	ctx, cancelTimeout := context.WithTimeout(context.Background(), time.Minute)
	defer cancelTimeout()

	// TODO clients benachrichtigen

	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("unable to shutdown server: %s", err)
	}

	cancel()

	log.Println("Shutting down Server")
}

// setUp sets up server handlers and the inactiveClientDeleter routine, which runs until the context cancels
func setUp(server *http.Server, handler *api.ServerHandler, timeLimit time.Duration, wg *sync.WaitGroup, ctx context.Context) {
	server.Handler = handler.BuildMultiplexer()

	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(15 * time.Second)

		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				handler.Service.InactiveClientDeleter(timeLimit)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ParseFlags parses server port, maximum users and tiemout duration flags
func ParseFlags() Config {
	var cfg Config

	flag.IntVar(&cfg.Port, "port", 8080, "HTTP Server Port")
	flag.IntVar(&cfg.maxUsers, "maxUsers", 100, "Maximum number of active users allowed")
	flag.DurationVar(&cfg.TimeLimit, "timeLimit", 30*time.Minute, "Time limit for inactive clients in minutes")
	flag.Parse()

	return cfg
}
