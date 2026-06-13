package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1"
	"github.com/KrishnaGrg1/auction_platform/internal/auth"
	db "github.com/KrishnaGrg1/auction_platform/internal/db/sqlc"
	"github.com/KrishnaGrg1/auction_platform/internal/helper"
	"github.com/KrishnaGrg1/auction_platform/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	store *store.Store
	jwt   *auth.JWTManager
}

func New(store *store.Store, jwt *auth.JWTManager) *Service {
	return &Service{
		store: store,
		jwt:   jwt,
	}
}
func (s *Service) Register(ctx context.Context, req *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error) {
	input := req.Msg
	existingUser, err := s.store.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser.Email != "" {
		return nil, helper.RpcError(connect.CodeAlreadyExists, "Email already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, err
	}

	newUser, err := s.store.Queries.CreateUser(ctx, db.CreateUserParams{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hashedPassword),
	})
	if err != nil {
		return nil, err
	}
	// generate a zero-padded 6-digit OTP using crypto/rand
	n, rErr := rand.Int(rand.Reader, big.NewInt(1000000))
	if rErr != nil {
		return nil, rErr
	}
	code := fmt.Sprintf("%06d", n.Int64())
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = s.store.Queries.CreateOTP(ctx, db.CreateOTPParams{
		UserID: newUser.ID,
		Code:   code,
		ExpiresAt: pgtype.Timestamptz{
			Time: expiresAt,
		},
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.RegisterResponse{
		User: &v1.User{
			Id:               newUser.ID.String(),
			Email:            newUser.Email,
			FirstName:        newUser.FirstName,
			LastName:         newUser.LastName,
			AvailableBalance: newUser.AvailableBalance,
			HeldBalance:      newUser.HeldBalance,
			IsVerified:       newUser.IsVerified,
			CreatedAt:        timestamppb.New(newUser.CreatedAt.Time),
			UpdatedAt:        timestamppb.New(newUser.UpdatedAt.Time),
		},
		Message: "Registered successfully",
	}), nil
}

func (s *Service) Login(ctx context.Context, req *connect.Request[v1.LoginRequest]) (*connect.Response[v1.LoginResponse], error) {
	input := req.Msg
	existingUser, err := s.store.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, helper.RpcError(connect.CodeNotFound, "invalid email or password")
	}
	if existingUser.Email == "" {
		return nil, helper.RpcError(connect.CodeNotFound, "User not exists")
	}
	if existingUser.IsVerified == false {
		return nil, helper.RpcError(connect.CodePermissionDenied, "Email not verified")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(input.Password)); err != nil {
		return nil, helper.RpcError(connect.CodeUnauthenticated, "Incorrect password")
	}
	token, err := s.jwt.GenerateToken(existingUser.ID.String(), existingUser.Email)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Fail to generate token")
	}
	return connect.NewResponse(&v1.LoginResponse{
		Token: token,
		User: &v1.User{
			Id:               existingUser.ID.String(),
			Email:            existingUser.Email,
			FirstName:        existingUser.FirstName,
			LastName:         existingUser.LastName,
			AvailableBalance: existingUser.AvailableBalance,
			HeldBalance:      existingUser.HeldBalance,
			IsVerified:       existingUser.IsVerified,
			CreatedAt:        timestamppb.New(existingUser.CreatedAt.Time),
			UpdatedAt:        timestamppb.New(existingUser.UpdatedAt.Time),
		},
		Message: "Login Sucessfully",
	}), nil

}

func (s *Service) VerifyUser(ctx context.Context, req *connect.Request[v1.VerifyRequest]) (*connect.Response[v1.VerifyResponse], error) {
	input := req.Msg
	var parsedUserId pgtype.UUID
	if err := parsedUserId.Scan(input.UserId); err != nil {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "invalid user id")
	}
	//1. GetUserById
	existingUser, err := s.store.Queries.GetUserByID(ctx, parsedUserId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "User not found")
	}
	//2. Get valid otp
	otp, err := s.store.Queries.GetValidOTPByUserId(ctx, parsedUserId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeNotFound, "OTP not found or expired")
	}
	// 3. check otp
	if otp.Code != input.Code {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "incorrect OTP")
	}
	// 4. verify User
	_, err = s.store.Queries.VerifyUser(ctx, existingUser.Email)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to verify User")
	}
	// 5. mark otp used
	err = s.store.Queries.MarkOTPUsed(ctx, parsedUserId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "failed to mark OTP used")
	}
	// 6. delete all otp which has been expired
	go func() {
		_ = s.store.Queries.DeleteExpiredOTPs(context.Background())
	}()

	// 7. generate token
	token, err := s.jwt.GenerateToken(input.UserId, existingUser.Email)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to generate token")
	}
	return connect.NewResponse(&v1.VerifyResponse{
			Token: token,
			User: &v1.User{
				Id:               existingUser.ID.String(),
				Email:            existingUser.Email,
				FirstName:        existingUser.FirstName,
				LastName:         existingUser.LastName,
				AvailableBalance: existingUser.AvailableBalance,
				HeldBalance:      existingUser.HeldBalance,
				IsVerified:       existingUser.IsVerified,
				CreatedAt:        timestamppb.New(existingUser.CreatedAt.Time),
				UpdatedAt:        timestamppb.New(existingUser.UpdatedAt.Time),
			},
			Message: "User verified successfully"}),
		nil
}
