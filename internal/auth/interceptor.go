package auth

import (
	"context"
	"strings"

	"connectrpc.com/connect"
)

type AuthInterceptor struct {
	jwtManager *JWTManager
}

func NewAuthInterceptor(jwtManager *JWTManager) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
	}
}

func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Skip authentication for health check
		if req.Spec().Procedure == "/health" {
			return next(ctx, req)
		}

		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, connect.NewError(connect.CodeUnauthenticated, nil)
		}

		tokenString := parts[1]
		claim, err := i.jwtManager.VerifyToken(tokenString)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		ctx = context.WithValue(ctx, UserIDKey, claim.UserID)
		ctx = context.WithValue(ctx, EmailKey, claim.Email)

		return next(ctx, req)
	}
}

func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
