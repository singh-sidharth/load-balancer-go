package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/singh-sidharth/load-balancer/internal/balancer"
	"github.com/singh-sidharth/load-balancer/internal/metrics"
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

	rawBackends := loadBackendsFromEnv()

	var servers []server.Server
	for _, raw := range rawBackends {
		srv, err := server.New(raw)
		if err != nil {
			log.Printf("failed to create backend %s: %v", raw, err)
			continue
		}
		servers = append(servers, srv)
	}

	// Fail if not backends were created.
	if len(servers) == 0 {
		log.Fatal("no valid backend servers configured")
	}
	// Register Prometheus metrics before starting the server.
	metrics.Register()

	// Perform an initial synchronous health check before serving traffic.
	//
	// This ensures backend health state is initialized before any request routing.
	// Without this, backends may be treated as healthy until the first background
	// health check runs, causing requests to be routed to unreachable servers.
	//
	// This prevents a startup-time correctness bug and makes the initialization
	// order explicit.

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

	mux := http.NewServeMux()

	//expose /metrics for Prometheus to scrape
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		srv, err := lb.NextServer()
		if err != nil {
			if errors.Is(err, balancer.ErrNoHealthyBackends) {
				http.Error(rec, "service unavailable", http.StatusServiceUnavailable)
			} else {
				http.Error(rec, "internal server error", http.StatusInternalServerError)
			}

			metrics.RequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(rec.statusCode),
				"none",
			).Inc()

			metrics.RequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(time.Since(start).Seconds())

			return
		}

		srv.Serve(rec, r)

		metrics.RequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(rec.statusCode),
			srv.Address(),
		).Inc()

		metrics.RequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(time.Since(start).Seconds())
	})

	serverAddr := os.Getenv("PORT")
	if serverAddr == "" {
		serverAddr = "8080"
	}

	// update with
	serverAddr = ":" + serverAddr

	log.Printf("load balancer listening on %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadBackendsFromEnv() []string {
	backendsEnv := os.Getenv("BACKENDS")
	if backendsEnv == "" {
		return []string{
			"http://localhost:8081",
			"http://localhost:8082",
			"http://localhost:8083",
		}
	}

	parts := strings.Split(backendsEnv, ",")
	var backends []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			backends = append(backends, trimmed)
		}
	}
	return backends
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
