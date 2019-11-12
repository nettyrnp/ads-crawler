package api

import (
	"github.com/gorilla/mux"
	"github.com/nettyrnp/ads-crawler/api/common"
	"github.com/nettyrnp/ads-crawler/api/middleware"

	"github.com/nettyrnp/ads-crawler/config"
)

func RunHTTP(c config.Config) {
	r := mux.NewRouter()

	r.Use(
		mux.CORSMethodMiddleware(r),
		mux.MiddlewareFunc(middleware.DefaultHeaders(c)),
		//mux.MiddlewareFunc(middleware.Debugger()),
		mux.MiddlewareFunc(middleware.RequestID()),
		mux.MiddlewareFunc(middleware.Logger(common.Logger)),
	)

	s := NewServer(r, c)
	api := &API{
		Config: c,
		Router: r,
		Server: s,
	}

	LoadModules(api)

	common.LogInfof("started HTTP server on %s\n", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		common.LogFatalf("starting HTTP server failed with %s", err)
	}
	// todo: graceful shutdown (with a log message)
}

func RunHTTPS(c config.Config) {
	r := mux.NewRouter()

	r.Use(
		mux.CORSMethodMiddleware(r),
		mux.MiddlewareFunc(middleware.DefaultHeaders(c)),
		//mux.MiddlewareFunc(middleware.Debugger()),
		mux.MiddlewareFunc(middleware.RequestID()),
		mux.MiddlewareFunc(middleware.Logger(common.Logger)),
	)

	s := NewServer(r, c)
	api := &API{
		Config: c,
		Router: r,
		Server: s,
	}

	LoadModules(api)

	//go func(){
	//	httpSrv := &http.Server{
	//		Addr:    "0.0.0.0:8081",
	//	}
	//	common.LogInfof("started HTTP server on %s\n", httpSrv.Addr)
	//	err := httpSrv.ListenAndServe()
	//	if err != nil {
	//		common.LogFatalf("starting HTTP server failed with %s", err)
	//	}
	//}()

	common.LogInfof("started HTTPS server on %s\n", s.Addr)
	err := s.ListenAndServeTLS(c.TransportPublicKey, c.TransportPrivateKey)
	//err := s.ListenAndServeTLS("./localhost+1.pem", "./localhost+1-key.pem")
	if err != nil {
		common.LogFatalf("starting HTTPS server failed with %s", err)
	}

}
