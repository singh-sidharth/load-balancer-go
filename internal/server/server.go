package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Server defines methods that a backend server must implement.
type Server interface {
	Address() string
	IsAlive() bool
	Serve(http.ResponseWriter, *http.Request)
}

// SimpleServer implements the Server interface and represents a backend.
type SimpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

// Address returns the address of the server.
func (s *SimpleServer) Address() string {
	return s.address
}

// IsAlive returns the health status (always true for simplicity).
func (s *SimpleServer) IsAlive() bool {
	return true
}

// Serve forwards the request to the backend using a reverse proxy.
func (s *SimpleServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}

// NewSimpleServer creates a SimpleServer instance.
func NewSimpleServer(address string) *SimpleServer {
	serverURL, err := url.Parse(address)
	if err != nil {
		panic(fmt.Sprintf("Error parsing server URL %s: %v", address, err))
	}

	return &SimpleServer{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serverURL),
	}
}
