package middleware

import (
	"net/http"
)

// Middleware type definition
type Middleware func(h http.Handler) http.Handler
