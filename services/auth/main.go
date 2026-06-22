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
	"github.com/KrishnaGrg1/auction_platform/internal/store"

	authHandler "github.com/KrishnaGrg1/auction_platform/services/auth/handler"
	authService "github.com/KrishnaGrg1/auction_platform/services/auth/service"
)

func main() {
	cfg := config.Load()

	s, err := store.Connect(cfg.DB_URL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer s.Pool.Close()

	jwtManager := auth.NewJWTManager(cfg.JWT_SECRET)

	// Initialize auth service
	svc := authService.New(s, jwtManager)
	authSvc := authHandler.New(svc)

	// Create HTTP mux
	mux := http.NewServeMux()
	// Register auth service handlers (no auth interceptor for auth service)
	path, handler := auction_platformv1connect.NewAuthServiceHandler(authSvc,
		connect.WithInterceptors(validate.NewInterceptor()),
	)
	mux.Handle(path, handler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":" + cfg.AUTH_PORT
	fmt.Printf("🔐 Auth service running on port: %s\n", cfg.AUTH_PORT)
	fmt.Printf("🔗 gRPC endpoint: http://localhost:%s\n", cfg.AUTH_PORT)
	fmt.Printf("📍 Endpoints:\n")
	fmt.Printf("   - POST /auction_platform.v1.AuthService/Register\n")
	fmt.Printf("   - POST /auction_platform.v1.AuthService/Login\n")
	fmt.Printf("   - POST /auction_platform.v1.AuthService/Verify\n")

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
