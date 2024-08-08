package main

import (
	"net/http"
	"strings"
)


func nyi(w http.ResponseWriter) {
	w.WriteHeader(501)
	w.Write([]byte("Not Yet Implemented"))
}

func trimPath(path string) string {
	return strings.TrimPrefix(path,"/secret")
}