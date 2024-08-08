package main

import (
	"log/slog"
	"net/http"
	"strings"
)

func (s *ApiServer) registerHandlers() {
	s.router.HandleFunc("/", s.root)
	s.router.HandleFunc("GET /secret/", s.readSecret)
	s.router.HandleFunc("POST /secret/", s.writeSecret)
	s.router.HandleFunc("DELETE /secret/", s.deleteSecret)
}

type ApiServer struct {
	router *http.ServeMux
	server *http.Server
	config *Config
	storage *Storage
}

func NewServer() *ApiServer {
	conf := InitConfig()
	return &ApiServer{http.NewServeMux(), &http.Server{}, conf, InitStorage(conf)}
}

func (s *ApiServer) root(w http.ResponseWriter, r *http.Request) {
	slog.Info("Hit root")
	w.Write([]byte(nil))
}

func (s *ApiServer) readSecret(w http.ResponseWriter, r *http.Request) {
	path := trimPath(r.URL.Path)
	s.storage.GetSecret(path, "a")
	nyi(w)
}

func (s *ApiServer) writeSecret(w http.ResponseWriter, r *http.Request) {
	path := trimPath(r.URL.Path)
	s.storage.StoreSecret(path, []byte("a"), []byte("a"), "a")
	nyi(w)
}

func (s *ApiServer) deleteSecret(w http.ResponseWriter, r *http.Request) {
	path := trimPath(r.URL.Path)
	s.storage.DeleteSecret(path, "a")
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