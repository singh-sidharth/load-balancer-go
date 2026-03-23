package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/singh-sidharth/load-balancer/internal/balancer"
	"github.com/singh-sidharth/load-balancer/internal/server"
)

func main() {
	// NOTE:
	// These are local backend servers used for testing the load balancer.
	// Each backend is expected to expose:
	//   - "/"        -> main handler
	//   - "/health"  -> health check endpoint (returns 200)
	//
	// If you don't have backends running, you can:
	// 1. Run simple local servers (see cmd/backend)
	// 2. Or temporarily replace these with public URLs (health checks will be limited)

	rawBackends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	var servers []server.Server
	for _, raw := range rawBackends {
		srv, err := server.New(raw)
		if err != nil {
			fmt.Printf("Error creating server for %s: %v\n", raw, err)
			continue
		}
		servers = append(servers, srv)
	}

	// Do an initial synchronous health check before serving traffic.
	// This avoids routing requests based on assumed backend state.
	healthClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	for _, srv := range servers {
		healthy := server.CheckHealth(srv, healthClient)
		log.Printf("initial health check: backend=%s healthy=%t", srv.Address(), healthy)
	}

	lb := balancer.New(servers)

	// Keep backend health state fresh in the background.
	go server.StartHealthCheck(servers, 5*time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv, err := lb.NextServer()
		if err != nil {
			if errors.Is(err, balancer.ErrNoHealthyBackends) {
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		srv.Serve(w, r)

	})

	serverAddr := os.Getenv("PORT")
	if serverAddr == "" {
		serverAddr = "8080"
	}

	// update with
	serverAddr = ":" + serverAddr

	log.Printf("load balancer listening on %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
