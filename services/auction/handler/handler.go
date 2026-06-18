package handler

import (
	"context"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1"
	"github.com/KrishnaGrg1/auction_platform/services/auction/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}
func (h *Handler) CreateAuction(ctx context.Context, req *connect.Request[v1.CreateAuctionRequest]) (*connect.Response[v1.CreateAuctionResponse], error) {
	auction, err := h.service.CreateAuction(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.CreateAuctionResponse{
		Auction: auction,
		Message: "Auction created successfully",
	}), nil
}

func (h *Handler) BidAuction(ctx context.Context, req *connect.Request[v1.BidAuctionRequest]) (*connect.Response[v1.BidAuctionResponse], error) {
	bid, err := h.service.BidAuction(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.BidAuctionResponse{
		Bid:       bid,
		Message:   "Bid placed successfully",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}

func (h *Handler) GetAuctionDetailsById(ctx context.Context, req *connect.Request[v1.GetAuctionDetailsByIdRequest]) (*connect.Response[v1.GetAuctionDetailsByIdResponse], error) {
	auction, err := h.service.GetAuctionDetailsById(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.GetAuctionDetailsByIdResponse{
		Auction:   auction,
		Message:   "Auction details retrieved successfully",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}

func (h *Handler) GetAuctionsList(ctx context.Context, req *connect.Request[v1.GetAuctionsListRequest]) (*connect.Response[v1.GetAuctionsListResponse], error) {
	auctions, err := h.service.GetAuctionsList(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.GetAuctionsListResponse{
		Auctions: auctions,
	}), nil
}

func (h *Handler) GetUserAuctions(ctx context.Context, req *connect.Request[v1.GetUserAuctionsRequest]) (*connect.Response[v1.GetUserAuctionsResponse], error) {
	auctions, err := h.service.GetUserAuctions(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.GetUserAuctionsResponse{
		Auctions: auctions,
	}), nil
}

func (h *Handler) EndAuction(ctx context.Context, req *connect.Request[v1.EndAuctionRequest]) (*connect.Response[v1.EndAuctionResponse], error) {
	auction, err := h.service.EndAuction(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.EndAuctionResponse{
		Auction:   auction,
		Message:   "End auction successfully",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}

func (h *Handler) CancelAuction(ctx context.Context, req *connect.Request[v1.CancelAuctionRequest]) (*connect.Response[v1.CancelAuctionResponse], error) {
	auction, err := h.service.CancelAuction(ctx, req)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.CancelAuctionResponse{
		Auction:   auction,
		Message:   "Cancelled auction successfully",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}
