package main

import (
	"io"
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
	token := r.Header.Get("X-Api-Key")
	path := trimPath(r.URL.Path)
	cipher, nonce := s.storage.GetSecret(path, token)
	if cipher != nil && nonce != nil {
		w.WriteHeader(200)
		w.Write(cipher)
		return
	}
	w.WriteHeader(500)
	w.Write([]byte("\"could not fetch secret\""))
}

func (s *ApiServer) writeSecret(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Api-Key")
	path := trimPath(r.URL.Path)
	plaintext, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("\"server error\""))
	}
	nonce := []byte("a")
	ok := s.storage.StoreSecret(path, plaintext, nonce, token)
	if ok {
		w.WriteHeader(200)
		w.Write([]byte("\"secret written\""))
	} else {
		w.WriteHeader(500)
		w.Write([]byte("\"secret not written\""))
	}
	
}

func (s *ApiServer) deleteSecret(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Api-Key")
	path := trimPath(r.URL.Path)
	ok := s.storage.DeleteSecret(path, token)
	if ok {
		w.WriteHeader(200)
		w.Write([]byte("\"secret deleted\""))
	} else {
		w.WriteHeader(500)
		w.Write([]byte("\"secret not deleted\""))
	}

}

func (s ApiServer) StartServer() {
	const port = 8080
	s.registerHandlers()
	slog.Info("Starting server", "Port", port)
	tokenRouter := http.NewServeMux()
	tokenRouter.Handle("/", s.checkTokenExists(s.router))
	logRouter := http.NewServeMux()
	logRouter.Handle("/", s.logHttpRequest(tokenRouter))
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

func (s * ApiServer) checkTokenExists(next http.Handler) (http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		token := r.Header.Get("X-Api-Key")
		if token == "" {
			w.WriteHeader(403)
			w.Write([]byte("Forbidden"))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}