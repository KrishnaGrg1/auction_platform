package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/KrishnaGrg1/auction_platform/internal/config"
	"github.com/KrishnaGrg1/auction_platform/internal/store"
)

func main() {
	cfg := config.Load()

	s, err := store.Connect(cfg.DB_URL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer s.Pool.Close()

	mux := http.NewServeMux()
	addr := ":" + cfg.PORT
	// Use new Go 1.23+ http.Protocols API for HTTP/2 support
	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true) // h2c - HTTP/2 without TLS

	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		Protocols: p,
	}
	fmt.Printf("server running on port: %s\n", cfg.PORT)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
