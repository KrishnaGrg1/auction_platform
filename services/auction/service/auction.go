package service

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1"
	"github.com/KrishnaGrg1/auction_platform/internal/auth"
	db "github.com/KrishnaGrg1/auction_platform/internal/db/sqlc"
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
func rpcError(code connect.Code, message string) error {
	return connect.NewError(code, errors.New(message))
}

func (s *Service) CreateAuction(ctx context.Context, req *connect.Request[v1.CreateAuctionRequest]) (*connect.Response[v1.CreateAuctionResponse], error) {
	userId := ctx.Value(auth.UserIDKey).(string)
	input := req.Msg
	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return nil, rpcError(connect.CodeInvalidArgument, "invalid user id")
	}
	seller_id := pgtype.UUID{
		Bytes: parsedUserId, // Copies the underlying [16]byte array natively
		Valid: true,
	}
	exisitingUser, err := s.store.Queries.GetUserByID(ctx, seller_id)
	if err != nil {
		return nil, rpcError(connect.CodeInternal, "User not found")
	}
	if exisitingUser.Email == "" {
		return nil, rpcError(connect.CodeInternal, "User not found")
	}

	// Build auction parameters based on type
	auctionParams := db.CreateAuctionParams{
		SellerID:      seller_id,
		Title:         input.Title,
		Description:   input.Description,
		Type:          db.AuctionType(input.Type.String()),
		StartingPrice: int32(input.StartingPrice),
		ReservedPrice: int32(input.ReservedPrice),
		ExtendOnBid:   input.ExtendOnBid,
		ExtendMinutes: input.ExtendMinutes,
		CurrentPrice:  int32(input.StartingPrice),
		StartTime:     pgtype.Timestamptz{Time: input.StartTime.AsTime(), Valid: true},
		EndTime:       pgtype.Timestamptz{Time: input.EndTime.AsTime(), Valid: true},
	}

	// For Dutch auction, add drop parameters
	if input.Type == v1.AuctionType_AUCTION_TYPE_DUTCH {
		if input.DropAmount <= 0 || input.DropInterval <= 0 {
			return nil, rpcError(connect.CodeInvalidArgument, "Dutch auction requires drop_amount and drop_interval")
		}
		auctionParams.DropAmount = pgtype.Int4{Int32: int32(input.DropAmount), Valid: true}
		auctionParams.DropInterval = pgtype.Int4{Int32: input.DropInterval, Valid: true}
	}
	newAuction, err := s.store.Queries.CreateAuction(ctx, auctionParams)
	if err != nil {
		return nil, rpcError(connect.CodeInternal, "Failed to create auction")
	}

	return connect.NewResponse(&v1.CreateAuctionResponse{
		Auction: &v1.Auction{
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
		},
		Message: "Auction created successfully",
	}), nil
}
func (s *Service) BidAuction(ctx context.Context, req *connect.Request[v1.BidAuctionRequest]) (*connect.Response[v1.BidAuctionResponse], error) {
	input := req.Msg
	userId := ctx.Value(auth.UserIDKey).(string)

	buyer_id, err := parsedStringToUUId(userId)
	if err != nil {
		return nil, err
	}
	auction_id, err := parsedStringToUUId(input.AuctionId)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, rpcError(connect.CodeInternal, "failed to start transaction")
	}
	defer tx.Rollback(ctx)

	qtx := s.store.Queries.WithTx(tx)

	// Lock user and auction for update
	_, err = qtx.LockUserByID(ctx, buyer_id)
	if err != nil {
		return nil, rpcError(connect.CodeInternal, "failed to lock user")
	}

	existingAuction, err := qtx.LockAuctionByID(ctx, auction_id)
	if err != nil {
		return nil, rpcError(connect.CodeNotFound, "Auction not found")
	}

	// Validate auction state
	if existingAuction.Status != db.AuctionStatusActive {
		return nil, rpcError(connect.CodeFailedPrecondition, "Auction not active")
	}
	if time.Now().After(existingAuction.EndTime.Time) {
		return nil, rpcError(connect.CodeFailedPrecondition, "Auction ended already")
	}
	if existingAuction.SellerID.Bytes == buyer_id.Bytes {
		return nil, rpcError(connect.CodeInvalidArgument, "Cannot bid on your own auction")
	}

	bidAmount := int32(input.Amount)

	// Validate bid amount
	if existingAuction.Type == db.AuctionTypeEnglish {
		// English auction: bid must be higher than current price
		if bidAmount <= existingAuction.CurrentPrice {
			return nil, rpcError(connect.CodeInvalidArgument, "Bid must be higher than current price")
		}
	} else {
		// Dutch auction: bid must be >= current price (accept current offer)
		if bidAmount < existingAuction.CurrentPrice {
			return nil, rpcError(connect.CodeInvalidArgument, "Bid must match or exceed current price")
		}
	}

	existingUser, err := qtx.GetUserByID(ctx, buyer_id)
	if err != nil {
		return nil, rpcError(connect.CodeNotFound, "User not found")
	}

	// Check if user already has an active bid on this auction
	existingBid, err := qtx.GetActiveBidByUserAndAuction(ctx, db.GetActiveBidByUserAndAuctionParams{
		AuctionID: auction_id,
		UserID:    buyer_id,
	})

	var newBid db.Bid
	if err == nil {
		// User is increasing their own bid
		bidDifference := bidAmount - existingBid.Amount
		if bidDifference <= 0 {
			return nil, rpcError(connect.CodeInvalidArgument, "New bid must be higher than your current bid")
		}
		if existingUser.AvailableBalance < bidDifference {
			return nil, rpcError(connect.CodeInvalidArgument, "Insufficient balance for bid increase")
		}

		// Hold the difference amount
		_, err = qtx.IncreaseHeldByDifference(ctx, db.IncreaseHeldByDifferenceParams{
			AvailableBalance: bidDifference,
			ID:               buyer_id,
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to hold bid amount")
		}

		// Create new bid (old one stays as outbid)
		newBid, err = qtx.CreateBid(ctx, db.CreateBidParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
			Amount:    bidAmount,
			IsAutoBid: input.IsAutoBid,
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to create bid")
		}

		// Mark old bid as outbid
		err = qtx.MarkBidOutbid(ctx, db.MarkBidOutbidParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to update previous bid")
		}
	} else {
		// New bidder
		if existingUser.AvailableBalance < bidAmount {
			return nil, rpcError(connect.CodeInvalidArgument, "Insufficient balance")
		}

		// Hold the full bid amount
		_, err = qtx.HoldBidAmount(ctx, db.HoldBidAmountParams{
			AvailableBalance: bidAmount,
			ID:               buyer_id,
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to hold bid amount")
		}

		// Create new bid
		newBid, err = qtx.CreateBid(ctx, db.CreateBidParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
			Amount:    bidAmount,
			IsAutoBid: input.IsAutoBid,
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to create bid")
		}

		// Release previous highest bidder's funds (if exists)
		if existingAuction.CurrentBidderID.Valid {
			prevBidderID := existingAuction.CurrentBidderID
			prevBid, err := qtx.GetActiveBidByUserAndAuction(ctx, db.GetActiveBidByUserAndAuctionParams{
				AuctionID: auction_id,
				UserID:    prevBidderID,
			})
			if err == nil {
				// Release previous bidder's held amount
				_, err = qtx.ReleaseBidAmount(ctx, db.ReleaseBidAmountParams{
					HeldBalance: prevBid.Amount,
					ID:          prevBidderID,
				})
				if err != nil {
					return nil, rpcError(connect.CodeInternal, "Failed to release previous bidder funds")
				}

				// Mark previous bid as outbid
				err = qtx.MarkBidOutbid(ctx, db.MarkBidOutbidParams{
					AuctionID: auction_id,
					UserID:    prevBidderID,
				})
				if err != nil {
					return nil, rpcError(connect.CodeInternal, "Failed to update previous bid status")
				}
			}
		}
	}

	// Update auction with new current price and bidder
	_, err = qtx.UpdateAuctionAfterBid(ctx, db.UpdateAuctionAfterBidParams{
		CurrentPrice: bidAmount,
		CurrentBidderID: pgtype.UUID{
			Bytes: buyer_id.Bytes,
			Valid: true,
		},
		ID: auction_id,
	})
	if err != nil {
		return nil, rpcError(connect.CodeInternal, "Failed to update auction")
	}

	// Record bid history
	historyEvent := db.AuctionEventBidPlaced
	historyNote := "New bid placed"
	if existingBid.ID.Valid {
		historyEvent = db.AuctionEventBidIncreased
		historyNote = "Bidder increased their bid"
	}

	_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
		AuctionID: auction_id,
		UserID:    buyer_id,
		Event:     historyEvent,
		Amount:    pgtype.Int4{Int32: bidAmount, Valid: true},
		Note:      pgtype.Text{String: historyNote, Valid: true},
	})
	if err != nil {
		return nil, rpcError(connect.CodeInternal, "Failed to create auction history")
	}

	// Dutch auction: End immediately after first bid
	if existingAuction.Type == db.AuctionTypeDutch {
		_, err = qtx.UpdateAuctionStatus(ctx, db.UpdateAuctionStatusParams{
			Status: db.AuctionStatusEnded,
			ID:     auction_id,
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to end Dutch auction")
		}

		// Mark winning bid as won
		err = qtx.MarkBidWon(ctx, newBid.ID)
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to mark bid as won")
		}

		// Record auction end in history
		_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
			AuctionID: auction_id,
			UserID:    buyer_id,
			Event:     db.AuctionEventEnded,
			Amount:    pgtype.Int4{Int32: bidAmount, Valid: true},
			Note:      pgtype.Text{String: "Dutch auction ended with first bid", Valid: true},
		})
		if err != nil {
			return nil, rpcError(connect.CodeInternal, "Failed to record auction end")
		}
	} else {
		// English auction: Extend auction time if configured
		if existingAuction.ExtendOnBid && existingAuction.ExtendMinutes > 0 {
			_, err = qtx.ExtendAuctionEndTime(ctx, db.ExtendAuctionEndTimeParams{
				Column1: existingAuction.ExtendMinutes,
				ID:      auction_id,
			})
			if err != nil {
				return nil, rpcError(connect.CodeInternal, "Failed to extend auction time")
			}

			// Record time extension in history
			_, err = qtx.CreateAuctionHistory(ctx, db.CreateAuctionHistoryParams{
				AuctionID: auction_id,
				UserID:    buyer_id,
				Event:     db.AuctionEventExtended,
				Amount:    pgtype.Int4{Int32: existingAuction.ExtendMinutes, Valid: true},
				Note:      pgtype.Text{String: "Auction time extended due to new bid", Valid: true},
			})
			if err != nil {
				return nil, rpcError(connect.CodeInternal, "Failed to record time extension")
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, rpcError(connect.CodeInternal, "Failed to commit transaction")
	}

	return connect.NewResponse(&v1.BidAuctionResponse{
		Bid: &v1.Bid{
			Id:        newBid.ID.String(),
			AuctionId: newBid.AuctionID.String(),
			UserId:    newBid.UserID.String(),
			Amount:    int64(newBid.Amount),
			Status:    v1.BidStatus(v1.BidStatus_value["BID_STATUS_"+string(newBid.Status)]),
			IsAutoBid: newBid.IsAutoBid,
			CreatedAt: timestamppb.New(newBid.CreatedAt.Time),
			UpdatedAt: timestamppb.New(newBid.UpdatedAt.Time),
		},
		Message:   "Bid placed successfully",
		Timestamp: timestamppb.New(time.Now()),
	}), nil
}

func parsedStringToUUId(s string) (pgtype.UUID, error) {
	parsedString, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, rpcError(connect.CodeInvalidArgument, "Cannot convert to uuid")
	}
	return pgtype.UUID{
		Bytes: parsedString, // Copies the underlying [16]byte array natively
		Valid: true,
	}, nil
}
