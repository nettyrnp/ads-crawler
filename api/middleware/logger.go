package middleware

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
)

func Logger(out io.Writer) Middleware {
	return func(h http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(out, h)
	}
}
