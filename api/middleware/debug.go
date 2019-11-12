package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nettyrnp/ads-crawler/api/common"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

func Debugger() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			x, err := httputil.DumpRequest(r, true)
			if err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
				return
			}
			common.LogInfo("\n\n" + string(x) + "\n")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)
			dump, err := httputil.DumpResponse(rec.Result(), false)
			if err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
				return
			}
			common.LogInfo("\n" + string(dump))
			// we copy the captured response headers to our new response
			for k, v := range rec.Header() {
				w.Header()[k] = v
			}

			// grab the captured response body
			data := rec.Body.Bytes()
			common.LogInfo(jsonPrettyPrint(string(data)))
			w.Write(data)
		})
	}
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "  ")
	if err != nil {
		return in
	}
	return out.String()
}
