package server_test

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/singh-sidharth/load-balancer/internal/balancer"
	"github.com/singh-sidharth/load-balancer/internal/server"
)

func getFreePort(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate free port: %v", err)
	}
	defer l.Close()

	return l.Addr().String()
}

func TestProxyFailureMarksBackendUnhealthyAndNextRequestFailsOver(t *testing.T) {
	// Set up a healthy backend server that responds to both "/" and "/health".
	healthyBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("healthy backend response"))
		}
	}))
	defer healthyBackend.Close()

	// Allocate an unused port and close it immediately so requests to it fail.
	deadAddr := getFreePort(t)
	deadBackendURL := "http://" + deadAddr

	deadBackend, err := server.New(deadBackendURL)
	if err != nil {
		t.Fatalf("failed to create dead backend: %v", err)
	}

	liveBackend, err := server.New(healthyBackend.URL)
	if err != nil {
		t.Fatalf("failed to create healthy backend: %v", err)
	}

	// Force both to appear healthy initially so round robin may choose the dead one.
	// The timout is 5 second before next health check, so this simulates a request arriving
	// before the health check detects the failure.
	deadBackend.SetAlive(true)
	liveBackend.SetAlive(true)

	lb := balancer.New([]server.Server{
		deadBackend,
		liveBackend,
	})

	// First selection should return the dead backend because it is first and marked alive.
	first, err := lb.NextServer()
	if err != nil {
		t.Fatalf("expected first backend selection to succeed, got err=%v", err)
	}

	if first.Address() != deadBackend.Address() {
		t.Fatalf("expected dead backend to be selected first, got %s", first.Address())
	}

	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "http://loadbalancer/", nil)

	first.Serve(rr1, req1)

	if rr1.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 from dead upstream, got %d", rr1.Code)
	}

	if deadBackend.IsAlive() {
		t.Fatal("expected dead backend to be marked unhealthy after proxy failure")
	}

	// Second request should now fail over to the healthy backend.
	second, err := lb.NextServer()
	if err != nil {
		t.Fatalf("expected second backend selection to succeed, got err=%v", err)
	}

	if second.Address() != liveBackend.Address() {
		t.Fatalf("expected healthy backend to be selected after failover, got %s", second.Address())
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "http://loadbalancer/", nil)

	second.Serve(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 from healthy backend, got %d", rr2.Code)
	}

	body, err := io.ReadAll(rr2.Result().Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Verify we got the expected response from the healthy backend, not a stale response from the dead one.
	// "strings.Contains" is used instead of exact match to avoid test fragility around response formatting.
	if !strings.Contains(string(body), "healthy backend response") {
		t.Fatalf("unexpected response body: %q", string(body))
	}
}

func TestNextServerReturnsErrorWhenAllBackendsBecomeUnavailable(t *testing.T) {
	deadAddr1 := getFreePort(t)
	deadAddr2 := getFreePort(t)

	s1, err := server.New("http://" + deadAddr1)
	if err != nil {
		t.Fatalf("failed to create backend 1: %v", err)
	}
	s2, err := server.New("http://" + deadAddr2)
	if err != nil {
		t.Fatalf("failed to create backend 2: %v", err)
	}

	s1.SetAlive(false)
	s2.SetAlive(false)

	lb := balancer.New([]server.Server{s1, s2})

	_, err = lb.NextServer()
	if !errors.Is(err, balancer.ErrNoHealthyBackends) {
		t.Fatalf("expected ErrNoHealthyBackends, got %v", err)
	}
}
