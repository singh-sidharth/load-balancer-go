package balancer

import (
	"net/http"
	"testing"

	"github.com/singh-sidharth/load-balancer/internal/server"
)

type fakeServer struct {
	address string
	alive   bool
}

func (f *fakeServer) Address() string {
	return f.address
}

func (f *fakeServer) IsAlive() bool {
	return f.alive
}

func (f *fakeServer) SetAlive(alive bool) {
	f.alive = alive
}

func (f *fakeServer) Serve(w http.ResponseWriter, r *http.Request) {
	// No-op for testing
}

func TestNextServer_RoundRobinAcrossHealthyBackends(t *testing.T) {
	// Arrange
	servers := []server.Server{
		&fakeServer{address: "a", alive: true},
		&fakeServer{address: "b", alive: true},
		&fakeServer{address: "c", alive: true},
	}

	lb := New(servers)

	//Act - The order matters here to verify round-robin behavior.
	got1, _ := lb.NextServer()
	got2, _ := lb.NextServer()
	got3, _ := lb.NextServer()
	got4, _ := lb.NextServer()

	// Assert
	if got1.Address() != "a" {
		t.Errorf("expected first server to be 'a', got '%s'", got1.Address())
	}

	if got2.Address() != "b" {
		t.Errorf("expected second server to be 'b', got '%s'", got2.Address())
	}

	if got3.Address() != "c" {
		t.Errorf("expected third server to be 'c', got '%s'", got3.Address())
	}

	if got4.Address() != "a" {
		t.Errorf("expected fourth server to wrap around to 'a', got '%s'", got4.Address())
	}
}

func TestNextServer_SkipsUnhealthyBackends(t *testing.T) {
	servers := []server.Server{
		&fakeServer{address: "a", alive: true},
		&fakeServer{address: "b", alive: false},
		&fakeServer{address: "c", alive: true},
	}

	lb := New(servers)

	got1, _ := lb.NextServer()
	got2, _ := lb.NextServer()
	got3, _ := lb.NextServer()

	if got1.Address() != "a" {
		t.Fatalf("expected a, got %s", got1.Address())
	}
	if got2.Address() != "c" {
		t.Fatalf("expected c, got %s", got2.Address())
	}
	if got3.Address() != "a" {
		t.Fatalf("expected a again, got %s", got3.Address())
	}
}

func TestNextServer_ReturnsErrorWhenNoHealthyBackends(t *testing.T) {
	servers := []server.Server{
		&fakeServer{address: "a", alive: false},
		&fakeServer{address: "b", alive: false},
	}

	lb := New(servers)

	_, err := lb.NextServer()
	if err != ErrNoHealthyBackends {
		t.Fatalf("expected ErrNoHealthyBackends, got %v", err)
	}
}
