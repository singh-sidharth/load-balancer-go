package balancer

import (
	"errors"
	"sync"

	"github.com/singh-sidharth/load-balancer/internal/server"
)

var ErrNoHealthyBackends = errors.New("no healthy backends available")

type LoadBalancer struct {
	servers []server.Server
	current int
	mu      sync.Mutex
}

func New(servers []server.Server) *LoadBalancer {
	return &LoadBalancer{
		servers: servers,
	}
}

func (lb *LoadBalancer) NextServer() (server.Server, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	n := len(lb.servers)
	if n == 0 {
		return nil, ErrNoHealthyBackends
	}

	start := lb.current

	for i := 0; i < n; i++ {
		idx := (start + i) % n
		srv := lb.servers[idx]
		if srv.IsAlive() {
			lb.current = (idx + 1) % n
			return srv, nil
		}
	}

	return nil, ErrNoHealthyBackends
}
