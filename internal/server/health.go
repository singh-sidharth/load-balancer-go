package server

import (
	"net/http"
	"time"
)

func CheckHealth(s Server, client *http.Client) bool {
	resp, err := client.Get(s.Address() + "/health")
	if err != nil {
		s.SetAlive(false)
		return false
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	s.SetAlive(healthy)

	return healthy
}

// This assumes each backend exposes GET /health
func StartHealthCheck(servers []Server, interval time.Duration) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		for _, srv := range servers {
			CheckHealth(srv, client)
		}
		<-ticker.C
	}
}
