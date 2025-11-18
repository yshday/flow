package middleware

import "net/http"

// Middleware is a function that wraps an HTTP handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares to a handler
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middlewares in reverse order so they execute in the order they're passed
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
