package main

import (
	"log/slog"
	"net/http"
	"strings"
)

func (s *ApiServer) registerHandlers() {
	s.router.HandleFunc("/", root)
	s.router.HandleFunc("GET /secret/{secretId}", readSecret)
	s.router.HandleFunc("POST /secret/{secretId}", writeSecret)
	s.router.HandleFunc("DELETE /secret/{secretId}", deleteSecret)
}

type ApiServer struct {
	router *http.ServeMux
	server *http.Server
}

func NewServer() *ApiServer {
	return &ApiServer{http.NewServeMux(), &http.Server{}}
}

func root(w http.ResponseWriter, r *http.Request) {
	slog.Info("Hit root")
	w.Write([]byte(nil))
}

func readSecret(w http.ResponseWriter, r *http.Request) {
	nyi(w)
}

func writeSecret(w http.ResponseWriter, r *http.Request) {
	nyi(w)
}

func deleteSecret(w http.ResponseWriter, r *http.Request) {
	nyi(w)
}

func (s ApiServer) StartServer() {
	const port = 8080
	s.registerHandlers()
	slog.Info("Starting server", "Port", port)
	logRouter := http.NewServeMux()
	logRouter.Handle("/", s.logHttpRequest(s.router))
	s.server = &http.Server{
		Addr : ":" + "8080",
		Handler: logRouter,
	}
	s.server.ListenAndServe()
}

func (s ApiServer) logHttpRequest(next http.Handler) (http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r * http.Request) {
		var remote string
		if r.Header.Get("X-Forwarded-For") != "" {
			remote = strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
		} else {
			remote = strings.Split(r.RemoteAddr, ":")[0]
		}
		
		slog.Info("", "RemoteAddr", remote, "Method" , r.Method, "Path" , r.URL.Path)
		next.ServeHTTP(w, r)
	})
}