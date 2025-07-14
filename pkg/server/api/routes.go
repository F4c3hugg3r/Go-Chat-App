package api

import "net/http"

// BuildMultiplexer adds handlers and middleware to the resulting multiplexer.
func (h *ServerHandler) BuildMultiplexer() http.Handler {
	multiplexer := http.NewServeMux()

	multiplexer.Handle("POST /users/{clientId}", http.HandlerFunc(h.HandleRegistry))
	multiplexer.Handle("POST /users/{clientId}/run", h.AuthMiddleware(h.HandleMessages))
	multiplexer.Handle("GET /users/{clientId}/chat", h.AuthMiddleware(h.HandleGetRequest))
	multiplexer.Handle("DELETE /users/{clientId}", h.AuthMiddleware(h.HandleMessages))

	return multiplexer
}
