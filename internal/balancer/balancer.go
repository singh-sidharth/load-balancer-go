package balancer

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/singh-sidharth/load-balancer/internal/server"
)

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []server.Server
	mu              sync.Mutex
}

// NewLoadBalancer creates a new LoadBalancer instance.
func NewLoadBalancer(port string, servers []server.Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

// getNextAvailableServer returns the next available server (round robin).
func (lb *LoadBalancer) getNextAvailableServer() server.Server {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i := 0; i < len(lb.servers); i++ {
		s := lb.servers[lb.roundRobinCount%len(lb.servers)]
		lb.roundRobinCount++
		if s.IsAlive() {
			return s
		}
	}
	if len(lb.servers) > 0 {
		return lb.servers[0]
	}
	return nil
}

// ServeProxy forwards the request to the next available backend.
func (lb *LoadBalancer) ServeProxy(rw http.ResponseWriter, r *http.Request) {
	target := lb.getNextAvailableServer()
	if target == nil {
		http.Error(rw, "No available servers", http.StatusServiceUnavailable)
		return
	}

	fmt.Printf("Forwarding request to %s\n", target.Address())
	target.Serve(rw, r)
}

// Port returns the load balancer port.
func (lb *LoadBalancer) Port() string {
	return lb.port
}
