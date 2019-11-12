package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nettyrnp/ads-crawler/config"
)

// NewServer creates a new api server instance
func NewServer(mux *mux.Router, config config.Config) *http.Server {
	return &http.Server{
		Addr:         config.Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}
