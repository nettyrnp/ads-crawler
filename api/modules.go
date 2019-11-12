package api

import (
	"github.com/nettyrnp/ads-crawler/api/sys"
	"github.com/nettyrnp/ads-crawler/api/sys/entity"
	"github.com/nettyrnp/ads-crawler/config"
)

// LoadModules of the api
func LoadModules(api *API) {
	api.Router = api.Router.PathPrefix("/api/v0").Subrouter()
	api.NewCrawlingModule(api.Config)
}

func (api *API) NewCrawlingModule(conf config.Config) {
	api.NewTPPModule(conf, entity.KindCrawler)
}

func (api *API) NewTPPModule(conf config.Config, kind0 entity.TPPKind) {
	kind := string(kind0)
	c := sys.NewController(conf, kind)
	sys.Route(api.Router, c)
}
