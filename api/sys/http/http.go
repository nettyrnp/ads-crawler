package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/nettyrnp/ads-crawler/api/sys/dto"
	"github.com/nettyrnp/ads-crawler/api/sys/entity"
	"github.com/nettyrnp/ads-crawler/api/sys/service"
	"github.com/nettyrnp/ads-crawler/config"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Controller struct {
	Kind    string
	Service service.Service
	Conf    config.Config
}

func New(s service.Service, conf config.Config, kind string) *Controller {
	return &Controller{
		Kind:    kind,
		Service: s,
		Conf:    conf,
	}
}

func (c *Controller) Version(w http.ResponseWriter, r *http.Request) {
	v := "0.0.1-1"
	w.Write([]byte(fmt.Sprintf("%s Service, version %s", c.Kind, v)))
}

func (c *Controller) Logs(w http.ResponseWriter, r *http.Request) {
	if c.Conf.AppEnv != config.AppEnvDev {
		common.LogError("Attempt to access logs in non-development mode")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	svcResp := dto.NewServiceResponse()
	log, err := common.GetLog(c.Conf)
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, err.Error())
		return
	}
	max := 5000
	if len(log) > max {
		log = "     <........... truncated ............>     \n" + log[len(log)-max:]
	}
	w.Write([]byte("Backend latest log: \n" + log))
}

func (c *Controller) StartPolling(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()
	start := time.Now()

	portals, err := c.Service.GetPortals(r.Context())
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "failed to find any portals in storage").Error())
		return
	}

	// todo: separate goroutins
	for _, portal := range portals {

		// Purge before inserting // todo: make purge + batch insert in one transaction
		err := c.Service.DeleteProvider(r.Context(), portal.CanonicalName)
		if err != nil {
			common.LogError(err.Error())
			c.respondNotOK(w, http.StatusBadRequest, svcResp, err.Error())
			return
		}

		url := portal.Protocol + "://" + portal.CanonicalName + "/ads.txt"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.respondNotOK(w, http.StatusInternalServerError, svcResp, err.Error())
			return
		}
		req.Header.Set("Content-Type", "plain/text; charset=utf-8")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			c.respondNotOK(w, http.StatusInternalServerError, svcResp, err.Error())
			return
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.respondNotOK(w, http.StatusInternalServerError, svcResp, err.Error())
			return
		}
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			bodyStr := string(body)
			lines := strings.SplitN(bodyStr, "\n", -1) // todo: os.Separator
			//n:= len(lines)
			providers := []*entity.Provider{}
			errs := []error{}
			for _, line := range lines {
				provider, parseErr := entity.ParseProvider(line)
				if parseErr != nil {
					if parseErr != entity.ErrInvalidLine { // log only specific parse errors
						errs = append(errs, parseErr)
					}
					continue
				}
				provider.PortalID = portal.ID
				provider.CreatedAt = time.Now().UTC()

				// Insert to storage
				id, err0 := c.Service.AddProvider(r.Context(), provider)
				if err0 != nil {
					errs = append(errs, err0)
					continue
				}
				fmt.Printf(">> Inserted id: %v\n", id)
				providers = append(providers, provider)
			}
			common.LogInfof("Got %v lines on portal '%v' [over %v]", len(lines), portal.CanonicalName, portal.Protocol)
			common.LogInfof("Parsed %v providers for portal '%v'", len(providers), portal.CanonicalName)
			if len(providers) > 0 {
				var sb bytes.Buffer
				for _, p := range providers {
					sb.WriteString(fmt.Sprintf("%v\n", *p))
				}
				common.LogInfof("Providers for portal '%v':\n%v", portal.CanonicalName, sb.String())
			}
			if len(errs) > 0 {
				common.LogErrorf("Errors for portal '%v':\n%v", portal.CanonicalName, common.JoinErrors(errs))
			}

		} else if res.StatusCode == http.StatusUnauthorized {
			// try https
			// ...

		} else if res.StatusCode == http.StatusNotFound {
			c.Service.NotifyPortalAdmins(r.Context(), portal)

		} else {
			common.LogErrorf("unexpected response code %v for portal %v", res.StatusCode, url) // todo: LogWarnf
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	common.LogInfof("Poll completed in %vs", (time.Now().Sub(start)).Seconds())
}

const (
	sortByDomain       = "domain"
	sortByCreationDate = "created"
)

func (c *Controller) GetPortals(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()

	portals, err := c.Service.GetPortals(r.Context())
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "failed to find any portals").Error())
		return
	}

	svcResp.Body = portals
	respondOK(w, svcResp, "")
}

func (c *Controller) GetPortalsExt(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()

	var req customerPortalsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errors.Wrap(err, "can't parse request body").Error())
		return
	}

	var sortBy entity.PortalSortField
	switch req.SortBy {
	case "", sortByDomain:
		sortBy = entity.SortByDomain
	case sortByCreationDate:
		sortBy = entity.SortByCreationDate
	default:
		errMsg := fmt.Sprintf("sorting attribute %s not supportred", req.SortBy)
		c.respondNotOK(w, http.StatusBadRequest, svcResp, errMsg)
		return
	}

	var from, till time.Time
	if req.From > 0 {
		from = time.Unix(req.From, 0)
	}
	if req.To > 0 {
		till = time.Unix(req.To, 0)
	}

	portals, total, err := c.Service.GetPortalsExt(r.Context(), from, till, sortBy, req.Desc, req.Limit, req.Offset)
	if err != nil {
		c.respondNotOK(w, http.StatusInternalServerError, svcResp, errors.Wrapf(err, "failed to find any portals").Error())
		return
	}

	svcResp.Body = &CustomerPortalsResp{
		Portals: portals,
		Total:   total,
	}
	respondOK(w, svcResp, "")
}

//func (c *Controller) GetPortals(w http.ResponseWriter, r *http.Request) {
//	svcResp := dto.NewServiceResponse()
//
//	opts:=repository.PortalsQueryOpts{
//		Limit:  0,
//		Offset: 0,
//	}
//	storedAds, lim, err := c.Service.GetPortals(r.Context(), opts)
//	if err != nil {
//		common.LogError(err.Error())
//		c.respondNotOK(w, http.StatusBadRequest, svcResp, err.Error())
//		return
//	}
//	if storedAds == nil {
//		c.respondNotOK(w, http.StatusUnauthorized, svcResp, "Wrong email or password")
//		return
//	}
//
//	svcResp.Body = *storedAds
//	common.LogInfof("Took ads for domain %v from storage", domain)
//	respondOK(w, svcResp, "")
//}

func (c *Controller) GetProvidersByPortal(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()
	portalName := mux.Vars(r)["name"]

	storedProviders, err := c.Service.GetProvidersByPortal(r.Context(), portalName)
	if err != nil {
		common.LogError(err.Error())
		c.respondNotOK(w, http.StatusBadRequest, svcResp, err.Error())
		return
	}

	svcResp.Body = storedProviders
	common.LogInfof("Retrieved %v providers for portalName '%v' from storage", len(storedProviders), portalName)
	respondOK(w, svcResp, "")
}

func (c *Controller) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	svcResp := dto.NewServiceResponse()
	portalName := mux.Vars(r)["name"]

	err := c.Service.DeleteProvider(r.Context(), portalName)
	if err != nil {
		common.LogError(err.Error())
		c.respondNotOK(w, http.StatusBadRequest, svcResp, err.Error())
		return
	}
	common.LogInfof("Deleted all providers for portalName '%v' from storage", portalName)
	respondOK(w, svcResp, "")
}

func (c *Controller) respondNotOK(w http.ResponseWriter, statusCode int, response *dto.ServiceResponse, errorMsg string) {
	if c.Conf.AppEnv == config.AppEnvDev {
		respondNotOKWithError(w, statusCode, response, errorMsg)
		return
	}
	common.LogError(errorMsg)
	respond(w, statusCode, response, "")
}

func respondNotOKWithError(w http.ResponseWriter, statusCode int, response *dto.ServiceResponse, errorMsg string) {
	common.LogError(errorMsg)
	respond(w, statusCode, response, errorMsg)
}

func respondOK(w http.ResponseWriter, response *dto.ServiceResponse, msg string) {
	if msg != "" {
		common.LogInfo(msg)
	}
	statusCode := http.StatusOK
	respond(w, statusCode, response, msg)
}

func respond(w http.ResponseWriter, statusCode int, response *dto.ServiceResponse, msg string) {
	response.Status.Code = statusCode
	response.Status.Text = msg
	jsonResponse, _ := json.Marshal(*response)
	w.WriteHeader(statusCode)
	w.Write(jsonResponse)
}
