package main

import (
	"net/http"
	"strings"
)

func nyi(w http.ResponseWriter) {
	w.WriteHeader(501)
	w.Write([]byte("\"not yet implemented\""))
}

func se(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("\"server error\""))
}

func trimPath(path string) string {
	return strings.TrimPrefix(path, "/secret")
}
