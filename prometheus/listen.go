package prometheus

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Run - Create Handle '/metrics'
func Run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
