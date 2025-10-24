package main

import (
	"fmt"
	"net/http"

	"github.com/singh-sidharth/load-balancer/internal/balancer"
	"github.com/singh-sidharth/load-balancer/internal/server"
)

func main() {
	servers := []server.Server{
		server.NewSimpleServer("https://www.facebook.com"),
		server.NewSimpleServer("http://www.bing.com"),
		server.NewSimpleServer("https://www.google.com"),
	}

	lb := balancer.NewLoadBalancer("8000", servers)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(rw, r)
	})

	fmt.Printf("Load balancer running on localhost:%s\n", lb.Port())
	http.ListenAndServe(":"+lb.Port(), nil)
}
