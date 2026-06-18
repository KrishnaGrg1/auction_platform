package handler

import (
	"context"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1"
	authService "github.com/KrishnaGrg1/auction_platform/services/auth/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	authService *authService.Service
}

func New(authService *authService.Service) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(ctx context.Context, req *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error) {
	user, otpCode, err := h.authService.Register(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.RegisterResponse{
		User:      user,
		Message:   "Registered successfully. Your verification code is: " + otpCode + " (expires in 7 days)",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}

func (h *Handler) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	token, user, err := h.authService.Login(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.LoginResponse{
		Token:     token,
		User:      user,
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}

func (h *Handler) Verify(ctx context.Context, req *connect.Request[v1.VerifyRequest]) (*connect.Response[v1.VerifyResponse], error) {
	token, user, err := h.authService.Verify(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.VerifyResponse{
		Token:     token,
		User:      user,
		Message:   "User verified successfully",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}
