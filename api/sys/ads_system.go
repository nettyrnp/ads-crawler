package sys

import (
	"github.com/nettyrnp/ads-crawler/api/sys/notify"
	"os"

	"github.com/gorilla/mux"

	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/nettyrnp/ads-crawler/api/sys/http"
	"github.com/nettyrnp/ads-crawler/api/sys/repository"
	"github.com/nettyrnp/ads-crawler/api/sys/service"
	"github.com/nettyrnp/ads-crawler/config"
)

func NewRepository(conf config.Config, kind string) *repository.RDBMSRepository {
	repo := &repository.RDBMSRepository{
		Name: kind,
		Cfg: repository.Config{
			Driver: conf.RepositoryDriver,
			DSN:    conf.RepositoryDSN,
		},
	}

	if initErr := repo.Init(); initErr != nil {
		common.LogError(initErr.Error())
		os.Exit(1)
	}

	return repo
}

func NewController(conf config.Config, kind string) *http.Controller {
	repo := NewRepository(conf, kind)
	ns := notify.NewSmsNotifier(conf, kind)
	emailNotifier := notify.NewEmailNotifier(conf, kind)

	svc := service.New(conf, kind, repo, ns, emailNotifier)

	return http.New(svc, conf, kind)
}

func Route(mux *mux.Router, c *http.Controller) {
	mux.HandleFunc("/crawler/admin/version", c.Version).Methods("GET")
	mux.HandleFunc("/crawler/admin/logs", c.Logs).Methods("GET")
	mux.HandleFunc("/crawler/start_poll", c.StartPolling).Methods("POST")

	mux.HandleFunc("/crawler/portals", c.GetPortals).Methods("GET", "OPTIONS")
	mux.HandleFunc("/crawler/portals", c.GetPortalsExt).Methods("POST", "OPTIONS")
	mux.HandleFunc("/crawler/providers/portal/{name}", c.GetProvidersByPortal).Methods("GET", "OPTIONS")

	mux.HandleFunc("/crawler/providers/portal/{name}", c.DeleteProvider).Methods("DELETE", "OPTIONS")
}
