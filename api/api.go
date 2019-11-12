package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nettyrnp/ads-crawler/config"
)

// API stucture
type API struct {
	Config config.Config
	Router *mux.Router
	Server *http.Server
}
