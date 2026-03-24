package server

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Server defines methods that a backend server must implement.
type Server interface {
	Address() string
	IsAlive() bool
	SetAlive(bool)
	Serve(http.ResponseWriter, *http.Request)
}

// // BackendServer represents an upstream backend managed by the load balancer.
type BackendServer struct {
	url   *url.URL
	proxy *httputil.ReverseProxy

	mu    sync.RWMutex
	alive bool
}

// New creates a BackendServer for the given upstream URL.
func New(rawURL string) (*BackendServer, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	s := &BackendServer{
		url:   parsedURL,
		alive: false,
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 1 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 2 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.Transport = transport

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		s.SetAlive(false)
		log.Printf("proxy error: backend=%s err=%v", s.Address(), err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	s.proxy = proxy

	return s, nil
}

// Address returns the address of the server.
func (s *BackendServer) Address() string {
	return s.url.String()
}

// IsAlive returns the health status.
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
