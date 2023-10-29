package main

import (
	"net/http"
)


func nyi(w http.ResponseWriter) {
	w.WriteHeader(501)
	w.Write([]byte("Not Yet Implemented"))
}