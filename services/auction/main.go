package main

import (
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1/auction_platformv1connect"
	"github.com/KrishnaGrg1/auction_platform/internal/auth"
	"github.com/KrishnaGrg1/auction_platform/internal/config"
	"github.com/KrishnaGrg1/auction_platform/internal/socket"
	"github.com/KrishnaGrg1/auction_platform/internal/store"
	auctionHandler "github.com/KrishnaGrg1/auction_platform/services/auction/handler"
	auctionService "github.com/KrishnaGrg1/auction_platform/services/auction/service"
)

func main() {
	cfg := config.Load()

	s, err := store.Connect(cfg.DB_URL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer s.Pool.Close()

	jwtManager := auth.NewJWTManager(cfg.JWT_SECRET)

	// Initialize auction service
	auctionSvc := auctionHandler.New(auctionService.New(s, jwtManager))

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register auction service handlers
	authInterceptor := auth.NewAuthInterceptor(jwtManager)
	path, handler := auction_platformv1connect.NewAuctionServiceHandler(
		auctionSvc,
		connect.WithInterceptors(validate.NewInterceptor()),
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(path, handler)

	// Register WebSocket endpoint
	mux.HandleFunc("/ws/auction", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(s.SocketHub(), w, r)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":" + cfg.AUCTION_PORT
	fmt.Printf("🚀 Auction service running on port: %s\n", cfg.AUCTION_PORT)
	fmt.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws/auction?auction_id=<auction_id>\n", cfg.AUCTION_PORT)
	fmt.Printf("🔗 gRPC endpoint: http://localhost:%s\n", cfg.AUCTION_PORT)

	// Use new Go 1.23+ HTTP/2 support
	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true) // h2c - HTTP/2 without TLS

	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		Protocols: p,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
