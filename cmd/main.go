package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/duncanleo/plex-dvr-hls/config"
	"github.com/duncanleo/plex-dvr-hls/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	var port = 5004
	var portStr = os.Getenv("PORT")
	var err error

	if len(portStr) > 0 {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	// serverCtx is cancelled on shutdown to signal active streams to kill their
	// ffmpeg processes before the process exits.
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/capability", routes.Capability)
	r.GET("/discover.json", routes.Discover)
	r.GET("/lineup.json", routes.Lineup)
	r.GET("/lineup_status.json", routes.LineupStatus)
	r.GET("/stream/:channelID", routes.StreamWithContext(serverCtx))
	r.GET("/xmltv", routes.XMLTV)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Starting '%s' tuner with encoder profile %s\n", config.Cfg.Name, config.Cfg.GetEncoderProfile())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal %v, shutting down\n", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v — shutting down\n", err)
	}

	// Cancel serverCtx first: signals active stream handlers to kill their
	// ffmpeg processes instead of waiting for them to drain naturally.
	serverCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown timed out (%v), forcing close\n", err)
		_ = srv.Close()
		os.Exit(1)
	}
	log.Println("Server stopped")
}
