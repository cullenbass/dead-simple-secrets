package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.Info("Starting dss...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := NewServer()
	defer s.storage.Cleanup()
	go s.StartServer()
	<-sig
	slog.Info("Closing, please wait...")
}
