package service

import (
	"context"
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

func (s *Service) CreateAuction(ctx context.Context, req *connect.Request[v1.CreateAuctionRequest]) (*connect.Response[v1.CreateAuctionResponse], error) {
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
			return nil, helper.RpcError(connect.CodeInvalidArgument, "Dutch auction requires drop_amount and drop_interval")
		}
		auctionParams.DropAmount = pgtype.Int4{Int32: int32(input.DropAmount), Valid: true}
		auctionParams.DropInterval = pgtype.Int4{Int32: input.DropInterval, Valid: true}
	}
	newAuction, err := s.store.Queries.CreateAuction(ctx, auctionParams)
	if err != nil {
		return nil, helper.RpcError(connect.CodeInternal, "Failed to create auction")
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
		timeLeft := existingAuction.EndTime.Time.Sub(time.Now())
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

func (s *Service) publishAuctionEvents(
	auction db.Auction,
	bid db.Bid,
	bidderID pgtype.UUID,
) {
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
