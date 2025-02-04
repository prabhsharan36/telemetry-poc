package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"telemetry/internal/config"
	"telemetry/internal/handlers"

	"github.com/fatih/color"
)

func abc() (int, error) {
	if f := 0; f == 0 {
		return 0, fmt.Errorf("math: square root of negative number %g", f)
	}

	return 0, nil
}

func main() {
	value, err := abc()

	if err != nil {
		fmt.Println(value, err)
	}

	config.LoadEnv()
	config.ConnectDatabase()
	// port := config.GetEnv("PORT", "8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/capture-event", handlers.TelemetryHandler)

	handler := loggingMiddleware(mux)
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		color.Green("✅ Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
