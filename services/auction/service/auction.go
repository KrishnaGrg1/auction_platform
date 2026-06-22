package service

import (
	"context"
	"math"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1"
	"github.com/KrishnaGrg1/auction_platform/internal/auth"
	db "github.com/KrishnaGrg1/auction_platform/internal/db/sqlc"
	"github.com/KrishnaGrg1/auction_platform/internal/helper"
	"github.com/KrishnaGrg1/auction_platform/internal/socket"
	"github.com/KrishnaGrg1/auction_platform/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (s *Service) CreateAuction(ctx context.Context, req *connect.Request[v1.CreateAuctionRequest]) (*v1.Auction, error) {
	userId := ctx.Value(auth.UserIDKey).(string)
	input := req.Msg
	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "invalid user id")
	}
	seller_id := pgtype.UUID{
		Bytes: parsedUserId, // Copies the underlying [16]byte array natively
		Valid: true,
	}
	existingUser, err := s.store.Queries.GetUserByID(ctx, seller_id)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "User not found")
	}
	if existingUser.Email == "" {
		return nil, helper.RpcError(connect.CodeInternal, "User not found")
	}

	// Convert auction type from proto enum to database enum
	var auctionType db.AuctionType
	switch input.Type {
	case v1.AuctionType_AUCTION_TYPE_ENGLISH:
		auctionType = db.AuctionTypeEnglish
	case v1.AuctionType_AUCTION_TYPE_DUTCH:
		auctionType = db.AuctionTypeDutch
	default:
		return nil, helper.RpcError(connect.CodeInvalidArgument, "Invalid auction type")
	}
	if input.Type == v1.AuctionType_AUCTION_TYPE_DUTCH {
		if input.DropAmount <= 0 || input.DropInterval <= 0 {
			return nil, helper.RpcError(connect.CodeInvalidArgument, "drop amount is less")
		}
	}

	if input.StartingPrice > math.MaxInt32 {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "starting_price too large")
	}
	// Build auction parameters based on type
	auctionParams := db.CreateAuctionParams{
		SellerID:      seller_id,
		Title:         input.Title,
		Description:   input.Description,
		Type:          auctionType,
		StartingPrice: int32(input.StartingPrice),
		ReservedPrice: int32(input.ReservedPrice),
		ExtendOnBid:   input.ExtendOnBid,
		ExtendMinutes: input.ExtendMinutes,
		CurrentPrice:  int32(input.StartingPrice),
		StartTime:     pgtype.Timestamptz{Time: input.StartTime.AsTime(), Valid: true},
		EndTime:       pgtype.Timestamptz{Time: input.EndTime.AsTime(), Valid: true},
		// Initialize as NULL for English auctions
		DropAmount:   pgtype.Int4{Valid: false},
		DropInterval: pgtype.Int4{Valid: false},
	}

	// For Dutch auction, add drop parameters
	if input.Type == v1.AuctionType_AUCTION_TYPE_DUTCH {
		if input.DropAmount <= 0 || input.DropInterval <= 0 {
			return nil, helper.RpcError(connect.CodeInvalidArgument, "Dutch auction requires drop_amount and drop_interval")
		}
		auctionParams.DropAmount = pgtype.Int4{Int32: int32(input.DropAmount), Valid: true}
		auctionParams.DropInterval = pgtype.Int4{Int32: int32(input.DropInterval), Valid: true}
	}
	newAuction, err := s.store.Queries.CreateAuction(ctx, auctionParams)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to create auction: "+err.Error())
	}

	return &v1.Auction{
		Id:              newAuction.ID.String(),
		SellerId:        uuid.UUID(newAuction.SellerID.Bytes).String(),
		Title:           newAuction.Title,
		Description:     newAuction.Description,
		Type:            v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(newAuction.Type)]),
		Status:          v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(newAuction.Status)]),
		StartingPrice:   int64(newAuction.StartingPrice),
		ReservedPrice:   int64(newAuction.ReservedPrice),
		CurrentPrice:    int64(newAuction.CurrentPrice),
		ExtendOnBid:     newAuction.ExtendOnBid,
		ExtendMinutes:   newAuction.ExtendMinutes,
		StartTime:       timestamppb.New(newAuction.StartTime.Time),
		EndTime:         timestamppb.New(newAuction.EndTime.Time),
		OriginalEndTime: timestamppb.New(newAuction.OriginalEndTime.Time),
		CreatedAt:       timestamppb.New(newAuction.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(newAuction.UpdatedAt.Time),
	}, nil
}
func (s *Service) BidAuction(ctx context.Context, req *connect.Request[v1.BidAuctionRequest]) (*v1.Bid, error) {
	input := req.Msg
	userId, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userId == "" {
		return nil, helper.RpcError(connect.CodeUnauthenticated, "Missing user authentication")
	}

	buyer_id, err := helper.ParsedStringToUUID(userId)
	if err != nil {
		return nil, err
	}
	auction_id, err := helper.ParsedStringToUUID(input.AuctionId)
	if err != nil {
		return nil, err
	}
	bidAmount := int32(input.Amount)
	if bidAmount <= 0 {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "Bid amount must be positive")
	}
	// transaction begin─────────────────────────────────────────────
	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to start transaction")
	}
	defer tx.Rollback(ctx) // 🔧 FIX: ensures locks always released on error

	qtx := s.store.Queries.WithTx(tx)

	// 🔒 Lock buyer + auction
	existingUser, err := qtx.LockUserByID(ctx, buyer_id)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to lock buyer")
	}
	existingAuction, err := qtx.LockAuctionByID(ctx, auction_id)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to lock auction")
	}

	// ── Common validations ──
	if existingAuction.SellerID.Bytes == buyer_id.Bytes {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "You cannot bid on your own auction")
	}
	if existingAuction.Status != db.AuctionStatusActive {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "Auction is not active")
	}
	if time.Now().After(existingAuction.EndTime.Time) {
		return nil, helper.RpcError(connect.CodeInvalidArgument, "Auction has ended")
	}

	var newBid db.Bid

	// ═══════════════════════════════════════════════
	// DUTCH AUCTION
	// ═══════════════════════════════════════════════
	if existingAuction.Type == db.AuctionTypeDutch {

		// 🔧 FIX: condition was backwards
		if bidAmount < existingAuction.CurrentPrice {
			return nil, helper.RpcError(connect.CodeInvalidArgument, "Bid amount is below current price")
		}

		// pay the CURRENT price (Dutch — not bidAmount, which may be higher)
		payAmount := existingAuction.CurrentPrice

		if existingUser.AvailableBalance < payAmount {
			return nil, helper.RpcError(connect.CodeInvalidArgument, "Insufficient balance to buy this item")
		}

		_, err = qtx.Withdraw(ctx, db.WithdrawParams{
			AvailableBalance: payAmount,
			ID:               buyer_id,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Cannot deduct amount from buyer")
		}

		_, err = qtx.Deposit(ctx, db.DepositParams{
			AvailableBalance: payAmount,
			ID:               existingAuction.SellerID,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Cannot credit seller")
		}

		// 🔧 FIX: use = not := (avoid shadowing outer newBid)
		newBid, err = qtx.CreateBid(ctx, db.CreateBidParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
			Amount:    payAmount,
			IsAutoBid: false,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to create bid")
		}

		err = qtx.MarkBidWon(ctx, newBid.ID)
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to mark bid as won")
		}

		_, err = qtx.CreateWinner(ctx, db.CreateWinnerParams{
			AuctionID:  auction_id,
			UserID:     buyer_id,
			FinalPrice: payAmount,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to create winner of auction")
		}
		newBid.Status = db.BidStatusWon
		_, err = qtx.UpdateAuctionStatus(ctx, db.UpdateAuctionStatusParams{
			Status: db.AuctionStatusEnded,
			ID:     auction_id,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to update auction status")
		}

		_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
			Event:     db.AuctionEventEnded,
			Amount:    pgtype.Int4{Int32: payAmount, Valid: true},
			Note:      pgtype.Text{String: "Dutch auction ended with first bid", Valid: true},
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to record auction end")
		}

		// ═══════════════════════════════════════════════
		// ENGLISH AUCTION
		// ═══════════════════════════════════════════════
	} else {

		if bidAmount <= existingAuction.CurrentPrice {
			return nil, helper.RpcError(connect.CodeInvalidArgument, "Bid amount must be higher than current price")
		}

		existingBid, err := qtx.GetActiveBidByUserAndAuction(ctx, db.GetActiveBidByUserAndAuctionParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
		})
		isIncrease := err == nil

		historyEvent := db.AuctionEventBidPlaced
		historyNote := "New bid placed"

		if isIncrease {
			// ── SAME BIDDER INCREASING ──
			bidDifference := bidAmount - existingBid.Amount
			if bidDifference <= 0 {
				return nil, helper.RpcError(connect.CodeInvalidArgument, "New bid must be higher than your current bid")
			}
			if existingUser.AvailableBalance < bidDifference {
				return nil, helper.RpcError(connect.CodeInvalidArgument, "Insufficient balance for bid increase")
			}

			_, err = qtx.HoldBidAmount(ctx, db.HoldBidAmountParams{
				AvailableBalance: bidDifference,
				ID:               buyer_id,
			})
			if err != nil {
				return nil, helper.RpcError(connect.CodeInternal, "Failed to hold bid amount")
			}

			err = qtx.MarkBidOutbid(ctx, db.MarkBidOutbidParams{
				AuctionID: auction_id,
				UserID:    buyer_id,
			})
			if err != nil { // 🔧 FIX: was unchecked before
				return nil, helper.RpcError(connect.CodeInternal, "Failed to mark old bid as outbid")
			}

			newBid, err = qtx.CreateBid(ctx, db.CreateBidParams{
				AuctionID: auction_id,
				UserID:    buyer_id,
				Amount:    bidAmount,
				IsAutoBid: input.IsAutoBid,
			})
			if err != nil {
				return nil, helper.RpcError(connect.CodeInternal, "Failed to create bid")
			}

			historyEvent = db.AuctionEventBidIncreased
			historyNote = "Bidder increased their bid"

		} else {
			// ── NEW / RETURNING BIDDER ──
			if existingUser.AvailableBalance < bidAmount {
				return nil, helper.RpcError(connect.CodeInvalidArgument, "Insufficient balance")
			}

			_, err = qtx.HoldBidAmount(ctx, db.HoldBidAmountParams{
				AvailableBalance: bidAmount,
				ID:               buyer_id,
			})
			if err != nil {
				return nil, helper.RpcError(connect.CodeInternal, "Failed to hold bid amount")
			}

			newBid, err = qtx.CreateBid(ctx, db.CreateBidParams{
				AuctionID: auction_id,
				Amount:    bidAmount,
				UserID:    buyer_id,
				IsAutoBid: input.IsAutoBid,
			})
			if err != nil {
				return nil, helper.RpcError(connect.CodeInternal, "Failed to create bid")
			}

			// 🔧 FIX: correct condition — does a previous bidder exist at all?
			if existingAuction.CurrentBidderID.Valid {
				prevBidderID := existingAuction.CurrentBidderID

				prevBid, err := qtx.GetActiveBidByUserAndAuction(ctx, db.GetActiveBidByUserAndAuctionParams{
					AuctionID: auction_id,
					UserID:    prevBidderID,
				})
				if err == nil {
					_, err = qtx.ReleaseBidAmount(ctx, db.ReleaseBidAmountParams{
						HeldBalance: prevBid.Amount,
						ID:          prevBidderID,
					})
					if err != nil {
						return nil, helper.RpcError(connect.CodeInternal, "Failed to release previous bidder funds")
					}

					err = qtx.MarkBidOutbid(ctx, db.MarkBidOutbidParams{
						AuctionID: auction_id,
						UserID:    prevBidderID,
					})
					if err != nil {
						return nil, helper.RpcError(connect.CodeInternal, "Failed to update previous bid status")
					}

					_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
						AuctionID: auction_id,
						UserID:    prevBidderID,
						Event:     db.AuctionEventBidRefunded,
						Amount:    pgtype.Int4{Int32: prevBid.Amount, Valid: true},
						Note:      pgtype.Text{String: "Outbid — funds released", Valid: true},
					})
					if err != nil {
						return nil, helper.RpcError(connect.CodeInternal, "Failed to record refund history")
					}
				}
			}
		}

		// 🔧 FIX: THIS WAS COMPLETELY MISSING — critical!
		_, err = qtx.UpdateAuctionAfterBid(ctx, db.UpdateAuctionAfterBidParams{
			CurrentPrice: bidAmount,
			CurrentBidderID: pgtype.UUID{
				Bytes: buyer_id.Bytes,
				Valid: true,
			},
			ID: auction_id,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to update auction")
		}

		// 🔧 FIX: was missing entirely
		_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
			Event:     historyEvent,
			Amount:    pgtype.Int4{Int32: bidAmount, Valid: true},
			Note:      pgtype.Text{String: historyNote, Valid: true},
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to create auction history")
		}

		// 🔧 FIX: anti-snipe extension was missing
		timeLeft := time.Until(existingAuction.EndTime.Time)
		extendWindow := time.Duration(existingAuction.ExtendMinutes) * time.Minute

		if existingAuction.ExtendOnBid && existingAuction.ExtendMinutes > 0 && timeLeft < extendWindow {
			_, err = qtx.ExtendAuctionEndTime(ctx, db.ExtendAuctionEndTimeParams{
				Column1: existingAuction.ExtendMinutes,
				ID:      auction_id,
			})
			if err != nil {
				return nil, helper.RpcError(connect.CodeInternal, "Failed to extend auction time")
			}

			_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
				AuctionID: auction_id,
				UserID:    buyer_id,
				Event:     db.AuctionEventExtended,
				Amount:    pgtype.Int4{Int32: existingAuction.ExtendMinutes, Valid: true},
				Note:      pgtype.Text{String: "Auction time extended due to late bid", Valid: true},
			})
			if err != nil {
				return nil, helper.RpcError(connect.CodeInternal, "Failed to record time extension")
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to commit transaction")
	}

	s.publishAuctionEvents(existingAuction, newBid, buyer_id)

	return &v1.Bid{
		Id:        newBid.ID.String(),
		AuctionId: newBid.AuctionID.String(),
		UserId:    newBid.UserID.String(),
		Amount:    int64(newBid.Amount),
		Status:    v1.BidStatus(v1.BidStatus_value["BID_STATUS_"+string(newBid.Status)]),
		IsAutoBid: newBid.IsAutoBid,
		CreatedAt: timestamppb.New(newBid.CreatedAt.Time),
		UpdatedAt: timestamppb.New(newBid.UpdatedAt.Time),
	}, nil
}

func (s *Service) GetAuctionDetailsById(ctx context.Context, req *connect.Request[v1.GetAuctionDetailsByIdRequest]) (*v1.Auction, error) {
	input := req.Msg
	auctionId, err := helper.ParsedStringToUUID(input.AuctionId)
	if err != nil {
		return nil, err
	}

	auction, err := s.store.Queries.GetAuctionByID(ctx, auctionId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeNotFound, "Auction not found")
	}

	return &v1.Auction{
		Id:       auction.ID.String(),
		SellerId: uuid.UUID(auction.SellerID.Bytes).String(),
		CurrentBidderId: func() string {
			if auction.CurrentBidderID.Valid {
				return uuid.UUID(auction.CurrentBidderID.Bytes).String()
			}
			return ""
		}(),
		Title:         auction.Title,
		Description:   auction.Description,
		Type:          v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(auction.Type)]),
		Status:        v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(auction.Status)]),
		StartingPrice: int64(auction.StartingPrice),
		ReservedPrice: int64(auction.ReservedPrice),
		CurrentPrice:  int64(auction.CurrentPrice),
		DropAmount: func() int64 {
			if auction.DropAmount.Valid {
				return int64(auction.DropAmount.Int32)
			}
			return 0
		}(),
		DropInterval: func() int32 {
			if auction.DropInterval.Valid {
				return auction.DropInterval.Int32
			}
			return 0
		}(),
		ExtendOnBid:     auction.ExtendOnBid,
		ExtendMinutes:   auction.ExtendMinutes,
		StartTime:       timestamppb.New(auction.StartTime.Time),
		EndTime:         timestamppb.New(auction.EndTime.Time),
		OriginalEndTime: timestamppb.New(auction.OriginalEndTime.Time),
		CreatedAt:       timestamppb.New(auction.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(auction.UpdatedAt.Time),
	}, nil

}

func (s *Service) GetAuctionsList(ctx context.Context, req *connect.Request[v1.GetAuctionsListRequest]) ([]*v1.Auction, error) {

	input := req.Msg

	var auctionType db.AuctionType
	switch input.Type {
	case v1.AuctionType_AUCTION_TYPE_ENGLISH:
		auctionType = db.AuctionTypeEnglish
	case v1.AuctionType_AUCTION_TYPE_DUTCH:
		auctionType = db.AuctionTypeDutch
	default:
		return nil, helper.RpcError(connect.CodeInvalidArgument, "Invalid auction type")
	}

	var status db.AuctionStatus
	switch input.Status {
	case v1.AuctionStatus_AUCTION_STATUS_ACTIVE:
		status = db.AuctionStatusActive
	case v1.AuctionStatus_AUCTION_STATUS_CANCELLED:
		status = db.AuctionStatusCancelled
	case v1.AuctionStatus_AUCTION_STATUS_ENDED:
		status = db.AuctionStatusEnded
	case v1.AuctionStatus_AUCTION_STATUS_SCHEDULED:
		status = db.AuctionStatusScheduled
	case v1.AuctionStatus_AUCTION_STATUS_UNSPECIFIED:
		status = db.AuctionStatusScheduled
	default:
		return nil, helper.RpcError(connect.CodeInvalidArgument, "Invalid auction status")
	}
	limt := input.Page
	skip := (input.Page - 1) * (input.PageSize)
	auctions, err := s.store.Queries.GetAuctionsList(ctx, db.GetAuctionsListParams{
		Status: status,
		Type:   auctionType,
		Limit:  limt,
		Offset: skip,
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to create auction: "+err.Error())
	}
	auctionList := make([]*v1.Auction, 0, len(auctions))
	for _, auction := range auctions {
		auctionList = append(auctionList, &v1.Auction{
			Id:       auction.ID.String(),
			SellerId: uuid.UUID(auction.SellerID.Bytes).String(),
			CurrentBidderId: func() string {
				if auction.CurrentBidderID.Valid {
					return uuid.UUID(auction.CurrentBidderID.Bytes).String()
				}
				return ""
			}(),
			Title:         auction.Title,
			Description:   auction.Description,
			Type:          v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(auction.Type)]),
			Status:        v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(auction.Status)]),
			StartingPrice: int64(auction.StartingPrice),
			ReservedPrice: int64(auction.ReservedPrice),
			CurrentPrice:  int64(auction.CurrentPrice),
			DropAmount: func() int64 {
				if auction.DropAmount.Valid {
					return int64(auction.DropAmount.Int32)
				}
				return 0
			}(),
			DropInterval: func() int32 {
				if auction.DropInterval.Valid {
					return auction.DropInterval.Int32
				}
				return 0
			}(),
			ExtendOnBid:     auction.ExtendOnBid,
			ExtendMinutes:   auction.ExtendMinutes,
			StartTime:       timestamppb.New(auction.StartTime.Time),
			EndTime:         timestamppb.New(auction.EndTime.Time),
			OriginalEndTime: timestamppb.New(auction.OriginalEndTime.Time),
			CreatedAt:       timestamppb.New(auction.CreatedAt.Time),
			UpdatedAt:       timestamppb.New(auction.UpdatedAt.Time),
		})
	}
	return auctionList, nil
}

func (s *Service) GetUserAuctions(ctx context.Context, req *connect.Request[v1.GetUserAuctionsRequest]) ([]*v1.Auction, error) {
	input := req.Msg
	userId, err := helper.ParsedStringToUUID(input.UserId)
	if err != nil {
		return nil, err
	}

	auctions, err := s.store.Queries.GetAuctionsBySellerID(ctx, userId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to retrieve user auctions")
	}

	auctionList := make([]*v1.Auction, 0, len(auctions))
	for _, auction := range auctions {
		auctionList = append(auctionList, &v1.Auction{
			Id:       auction.ID.String(),
			SellerId: uuid.UUID(auction.SellerID.Bytes).String(),
			CurrentBidderId: func() string {
				if auction.CurrentBidderID.Valid {
					return uuid.UUID(auction.CurrentBidderID.Bytes).String()
				}
				return ""
			}(),
			Title:         auction.Title,
			Description:   auction.Description,
			Type:          v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(auction.Type)]),
			Status:        v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(auction.Status)]),
			StartingPrice: int64(auction.StartingPrice),
			ReservedPrice: int64(auction.ReservedPrice),
			CurrentPrice:  int64(auction.CurrentPrice),
			DropAmount: func() int64 {
				if auction.DropAmount.Valid {
					return int64(auction.DropAmount.Int32)
				}
				return 0
			}(),
			DropInterval: func() int32 {
				if auction.DropInterval.Valid {
					return auction.DropInterval.Int32
				}
				return 0
			}(),
			ExtendOnBid:     auction.ExtendOnBid,
			ExtendMinutes:   auction.ExtendMinutes,
			StartTime:       timestamppb.New(auction.StartTime.Time),
			EndTime:         timestamppb.New(auction.EndTime.Time),
			OriginalEndTime: timestamppb.New(auction.OriginalEndTime.Time),
			CreatedAt:       timestamppb.New(auction.CreatedAt.Time),
			UpdatedAt:       timestamppb.New(auction.UpdatedAt.Time),
		})
	}

	return auctionList, nil
}

func (s *Service) CancelAuction(ctx context.Context, req *connect.Request[v1.CancelAuctionRequest]) (*v1.Auction, error) {
	input := req.Msg

	userId, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userId == "" {
		return nil, helper.RpcError(connect.CodeUnauthenticated, "Missing user authentication")
	}

	sellerId, err := helper.ParsedStringToUUID(userId)
	if err != nil {
		return nil, err
	}
	auctionId, err := helper.ParsedStringToUUID(input.AuctionId)
	if err != nil {
		return nil, err
	}

	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to start transaction")
	}
	defer tx.Rollback(ctx)

	qtx := s.store.Queries.WithTx(tx)

	existingAuction, err := qtx.LockAuctionByID(ctx, auctionId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeNotFound, "Auction not found")
	}

	if existingAuction.SellerID.Bytes != sellerId.Bytes {
		return nil, helper.RpcError(connect.CodePermissionDenied, "Only seller can cancel the auction")
	}

	if existingAuction.Status != db.AuctionStatusActive && existingAuction.Status != db.AuctionStatusScheduled {
		return nil, helper.RpcError(connect.CodeFailedPrecondition, "Auction cannot be cancelled")
	}

	canceledAuction, err := qtx.CancelAuction(ctx, db.CancelAuctionParams{
		ID:       auctionId,
		SellerID: sellerId,
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to cancel auction")
	}

	if existingAuction.CurrentBidderID.Valid {
		_, err = qtx.ReleaseBidAmount(ctx, db.ReleaseBidAmountParams{
			HeldBalance: existingAuction.CurrentPrice,
			ID:          existingAuction.CurrentBidderID,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to refund bidder")
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to commit transaction")
	}

	s.publishAuctionEvents(canceledAuction, db.Bid{}, sellerId)

	return &v1.Auction{
		Id:              canceledAuction.ID.String(),
		SellerId:        uuid.UUID(canceledAuction.SellerID.Bytes).String(),
		Title:           canceledAuction.Title,
		Description:     canceledAuction.Description,
		Type:            v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(canceledAuction.Type)]),
		Status:          v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(canceledAuction.Status)]),
		StartingPrice:   int64(canceledAuction.StartingPrice),
		ReservedPrice:   int64(canceledAuction.ReservedPrice),
		CurrentPrice:    int64(canceledAuction.CurrentPrice),
		ExtendOnBid:     canceledAuction.ExtendOnBid,
		ExtendMinutes:   canceledAuction.ExtendMinutes,
		StartTime:       timestamppb.New(canceledAuction.StartTime.Time),
		EndTime:         timestamppb.New(canceledAuction.EndTime.Time),
		OriginalEndTime: timestamppb.New(canceledAuction.OriginalEndTime.Time),
		CreatedAt:       timestamppb.New(canceledAuction.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(canceledAuction.UpdatedAt.Time),
	}, nil
}

func (s *Service) publishAuctionEvents(auction db.Auction, bid db.Bid, bidderID pgtype.UUID) {
	if s.store.SocketHub() == nil {
		return
	}

	switch auction.Type {

	case db.AuctionTypeDutch:

		s.store.SocketHub().BroadcastToAuction(
			auction.ID.String(),
			socket.AuctionEvent{
				Type:      "auction_won",
				AuctionID: auction.ID.String(),
				UserID:    uuid.UUID(bidderID.Bytes).String(),
				Amount:    int64(bid.Amount),
				Timestamp: time.Now(),
			},
		)

		s.store.SocketHub().BroadcastToAuction(
			auction.ID.String(),
			socket.AuctionEvent{
				Type:      "auction_ended",
				AuctionID: auction.ID.String(),
				Timestamp: time.Now(),
			},
		)

	default:

		s.store.SocketHub().BroadcastToAuction(
			auction.ID.String(),
			socket.AuctionEvent{
				Type:      "new_bid",
				AuctionID: auction.ID.String(),
				UserID:    uuid.UUID(bidderID.Bytes).String(),
				Amount:    int64(bid.Amount),
				Timestamp: time.Now(),
			},
		)
	}
}

func (s *Service) EndAuction(ctx context.Context, req *connect.Request[v1.EndAuctionRequest]) (*v1.Auction, error) {
	input := req.Msg

	userId, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userId == "" {
		return nil, helper.RpcError(connect.CodeUnauthenticated, "Missing user authentication")
	}

	sellerId, err := helper.ParsedStringToUUID(userId)
	if err != nil {
		return nil, err
	}
	auctionId, err := helper.ParsedStringToUUID(input.AuctionId)
	if err != nil {
		return nil, err
	}

	// ─── begin transaction ───────────────────────────────
	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to start transaction")
	}
	defer tx.Rollback(ctx)
	qtx := s.store.Queries.WithTx(tx)

	// 🔒 lock auction
	existingAuction, err := qtx.LockAuctionByID(ctx, auctionId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeNotFound, "Auction not found")
	}

	// ─── validations ────────────────────────────────────
	// 🔧 FIX: correct type comparison
	if existingAuction.SellerID.Bytes != sellerId.Bytes {
		return nil, helper.RpcError(connect.CodePermissionDenied, "Only seller can end the auction")
	}

	// 🔧 FIX: must be Active to end
	if existingAuction.Status != db.AuctionStatusActive {
		return nil, helper.RpcError(connect.CodeFailedPrecondition, "Auction is not active")
	}

	// ─── case 1: nobody bid ──────────────────────────────
	if !existingAuction.CurrentBidderID.Valid {
		// no bids at all — just cancel
		_, err = qtx.UpdateAuctionStatus(ctx, db.UpdateAuctionStatusParams{
			Status: db.AuctionStatusCancelled,
			ID:     auctionId,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to cancel auction")
		}

		_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
			AuctionID: auctionId,
			UserID:    sellerId,
			Event:     db.AuctionEventCancelled,
			Amount:    pgtype.Int4{Valid: false},
			Note:      pgtype.Text{String: "Auction ended with no bids", Valid: true},
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to record auction history")
		}

		if err = tx.Commit(ctx); err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to commit transaction")
		}

		cancelledAuction, err := s.store.Queries.GetAuctionByID(ctx, auctionId)
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to get updated auction")
		}

		s.publishAuctionEvents(cancelledAuction, db.Bid{}, sellerId)

		return &v1.Auction{
			Id:              cancelledAuction.ID.String(),
			SellerId:        uuid.UUID(cancelledAuction.SellerID.Bytes).String(),
			Title:           cancelledAuction.Title,
			Description:     cancelledAuction.Description,
			Type:            v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(cancelledAuction.Type)]),
			Status:          v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(cancelledAuction.Status)]),
			StartingPrice:   int64(cancelledAuction.StartingPrice),
			ReservedPrice:   int64(cancelledAuction.ReservedPrice),
			CurrentPrice:    int64(cancelledAuction.CurrentPrice),
			ExtendOnBid:     cancelledAuction.ExtendOnBid,
			ExtendMinutes:   cancelledAuction.ExtendMinutes,
			StartTime:       timestamppb.New(cancelledAuction.StartTime.Time),
			EndTime:         timestamppb.New(cancelledAuction.EndTime.Time),
			OriginalEndTime: timestamppb.New(cancelledAuction.OriginalEndTime.Time),
			CreatedAt:       timestamppb.New(cancelledAuction.CreatedAt.Time),
			UpdatedAt:       timestamppb.New(cancelledAuction.UpdatedAt.Time),
		}, nil
	}

	// ─── case 2: reserve price not met ──────────────────
	if existingAuction.CurrentPrice < existingAuction.ReservedPrice {

		// 🔒 lock current (highest) bidder to refund them
		_, err = qtx.LockUserByID(ctx, existingAuction.CurrentBidderID)
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to lock bidder")
		}

		// refund highest bidder's held money back to available
		_, err = qtx.ReleaseBidAmount(ctx, db.ReleaseBidAmountParams{
			HeldBalance: existingAuction.CurrentPrice,
			ID:          existingAuction.CurrentBidderID,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to refund bidder")
		}

		// mark their bid as refunded
		err = qtx.MarkBidOutbid(ctx, db.MarkBidOutbidParams{
			AuctionID: auctionId,
			UserID:    existingAuction.CurrentBidderID,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to update bid status")
		}

		_, err = qtx.UpdateAuctionStatus(ctx, db.UpdateAuctionStatusParams{
			Status: db.AuctionStatusNoReserve,
			ID:     auctionId,
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to update auction status")
		}

		_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
			AuctionID: auctionId,
			UserID:    existingAuction.CurrentBidderID,
			Event:     db.AuctionEventCancelled,
			Amount:    pgtype.Int4{Int32: existingAuction.CurrentPrice, Valid: true},
			Note:      pgtype.Text{String: "Reserve price not met — bidder refunded", Valid: true},
		})
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to record auction history")
		}

		if err = tx.Commit(ctx); err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to commit transaction")
		}

		noReserveAuction, err := s.store.Queries.GetAuctionByID(ctx, auctionId)
		if err != nil {
			return nil, helper.RpcError(connect.CodeInternal, "Failed to get updated auction")
		}

		s.publishAuctionEvents(noReserveAuction, db.Bid{}, existingAuction.CurrentBidderID)

		return &v1.Auction{
			Id:              noReserveAuction.ID.String(),
			SellerId:        uuid.UUID(noReserveAuction.SellerID.Bytes).String(),
			Title:           noReserveAuction.Title,
			Description:     noReserveAuction.Description,
			Type:            v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(noReserveAuction.Type)]),
			Status:          v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(noReserveAuction.Status)]),
			StartingPrice:   int64(noReserveAuction.StartingPrice),
			ReservedPrice:   int64(noReserveAuction.ReservedPrice),
			CurrentPrice:    int64(noReserveAuction.CurrentPrice),
			ExtendOnBid:     noReserveAuction.ExtendOnBid,
			ExtendMinutes:   noReserveAuction.ExtendMinutes,
			StartTime:       timestamppb.New(noReserveAuction.StartTime.Time),
			EndTime:         timestamppb.New(noReserveAuction.EndTime.Time),
			OriginalEndTime: timestamppb.New(noReserveAuction.OriginalEndTime.Time),
			CreatedAt:       timestamppb.New(noReserveAuction.CreatedAt.Time),
			UpdatedAt:       timestamppb.New(noReserveAuction.UpdatedAt.Time),
		}, nil
	}

	// ─── case 3: normal win ──────────────────────────────

	// 🔒 lock winner's user row before touching their balance
	_, err = qtx.LockUserByID(ctx, existingAuction.CurrentBidderID)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to lock winner")
	}

	// move winner's held money → 0 (they paid)
	_, err = qtx.TransferHeldToAvailable(ctx, db.TransferHeldToAvailableParams{
		HeldBalance: existingAuction.CurrentPrice,
		ID:          existingAuction.CurrentBidderID, // 🔧 FIX: winner user ID not auction ID
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to transfer winner funds")
	}

	// pay the seller
	_, err = qtx.CreditAvailableBalance(ctx, db.CreditAvailableBalanceParams{
		AvailableBalance: existingAuction.CurrentPrice,
		ID:               sellerId,
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to pay seller")
	}

	// get the winning bid row
	winningBid, err := qtx.GetHighestBidByAuctionID(ctx, auctionId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to get winning bid")
	}

	// mark winning bid as Won
	err = qtx.MarkBidWon(ctx, winningBid.ID)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to mark bid as won")
	}

	// mark all other outbid bids as Lost (cleanup)
	err = qtx.MarkAllLosingBids(ctx, auctionId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to mark losing bids")
	}

	// create winner record
	_, err = qtx.CreateWinner(ctx, db.CreateWinnerParams{
		AuctionID:  auctionId,
		UserID:     existingAuction.CurrentBidderID,
		FinalPrice: existingAuction.CurrentPrice,
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to create winner record")
	}

	// update auction status
	_, err = qtx.UpdateAuctionStatus(ctx, db.UpdateAuctionStatusParams{
		Status: db.AuctionStatusEnded,
		ID:     auctionId,
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to update auction status")
	}

	// 🔧 FIX: correct event (was AuctionEventExtended — wrong!)
	_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
		AuctionID: auctionId,
		UserID:    existingAuction.CurrentBidderID,
		Event:     db.AuctionEventWinnerDeclared,
		Amount:    pgtype.Int4{Int32: existingAuction.CurrentPrice, Valid: true},
		Note:      pgtype.Text{String: "Auction ended — winner declared", Valid: true},
	})
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to record auction history")
	}

	// 🔧 FIX: commit was missing entirely!
	if err = tx.Commit(ctx); err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to commit transaction")
	}

	// Fetch updated auction
	endedAuction, err := s.store.Queries.GetAuctionByID(ctx, auctionId)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to get updated auction")
	}

	s.publishAuctionEvents(endedAuction, winningBid, existingAuction.CurrentBidderID)

	return &v1.Auction{
		Id:              endedAuction.ID.String(),
		SellerId:        uuid.UUID(endedAuction.SellerID.Bytes).String(),
		CurrentBidderId: uuid.UUID(endedAuction.CurrentBidderID.Bytes).String(),
		Title:           endedAuction.Title,
		Description:     endedAuction.Description,
		Type:            v1.AuctionType(v1.AuctionType_value["AUCTION_TYPE_"+string(endedAuction.Type)]),
		Status:          v1.AuctionStatus(v1.AuctionStatus_value["AUCTION_STATUS_"+string(endedAuction.Status)]),
		StartingPrice:   int64(endedAuction.StartingPrice),
		ReservedPrice:   int64(endedAuction.ReservedPrice),
		CurrentPrice:    int64(endedAuction.CurrentPrice),
		ExtendOnBid:     endedAuction.ExtendOnBid,
		ExtendMinutes:   endedAuction.ExtendMinutes,
		StartTime:       timestamppb.New(endedAuction.StartTime.Time),
		EndTime:         timestamppb.New(endedAuction.EndTime.Time),
		OriginalEndTime: timestamppb.New(endedAuction.OriginalEndTime.Time),
		CreatedAt:       timestamppb.New(endedAuction.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(endedAuction.UpdatedAt.Time),
	}, nil
}

func (s *Service) GetMe(ctx context.Context, req *connect.Request[v1.GetMeRequest]) (*v1.User, error) {
	userId, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok || userId == "" {
		return nil, helper.RpcError(connect.CodeUnauthenticated, "Missing user authentication")
	}
	parsedUserId, err := helper.ParsedStringToUUID(userId)
	if err != nil {
		return nil, err
	}
	existingUser, err := s.store.Queries.GetUserByID(ctx, parsedUserId)
	if err != nil {
		return nil, err
	}
	user := &v1.User{
		Id:               existingUser.ID.String(),
		Email:            existingUser.Email,
		FirstName:        existingUser.FirstName,
		LastName:         existingUser.LastName,
		AvailableBalance: existingUser.AvailableBalance,
		HeldBalance:      existingUser.HeldBalance,
		IsVerified:       existingUser.IsVerified,
		CreatedAt:        timestamppb.New(existingUser.CreatedAt.Time),
		UpdatedAt:        timestamppb.New(existingUser.UpdatedAt.Time),
	}
	return user, nil
}
