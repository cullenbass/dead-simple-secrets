package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.Info("Starting dss...")
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := NewServer()
	go s.StartServer()
	<- sig
	slog.Info("Closing, please wait...")
}