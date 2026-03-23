package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// Server defines methods that a backend server must implement.
type Server interface {
	Address() string
	IsAlive() bool
	SetAlive(bool)
	Serve(http.ResponseWriter, *http.Request)
}

// Backend server represents managed backend servers in the load balancer.
type BackendServer struct {
	url   *url.URL
	proxy *httputil.ReverseProxy

	mu    sync.RWMutex
	alive bool
}

// Returns a new instance of BackendServer.
func New(rawURL string) (*BackendServer, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &BackendServer{
		url:   parsedURL,
		proxy: httputil.NewSingleHostReverseProxy(parsedURL),
		alive: true,
	}, nil
}

// Address returns the address of the server.
func (s *BackendServer) Address() string {
	return s.url.String()
}

// IsAlive returns the health status
func (s *BackendServer) IsAlive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alive
}

// SetAlive sets the health status of the server.
func (s *BackendServer) SetAlive(alive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alive = alive
}

// Serve forwards the request to the backend using a reverse proxy.
func (s *BackendServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}
