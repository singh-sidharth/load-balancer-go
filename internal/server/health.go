package server

import (
	"net/http"
	"time"
)

func CheckHealth(s Server, client *http.Client) bool {
	req, err := http.NewRequest(http.MethodGet, s.Address()+"/health", nil)
	if err != nil {
		s.SetAlive(false)
		return false
	}

	resp, err := client.Do(req)
	// Health check failed because the backend is unreachable due to network error or timeout.
	if err != nil {
		s.SetAlive(false)
		return false
	}
	defer resp.Body.Close()

	// Backend is reachable, but health status is unhealthy if it returns non-2xx.
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

	for range ticker.C {
		for _, srv := range servers {
			CheckHealth(srv, client)
		}
	}
}
