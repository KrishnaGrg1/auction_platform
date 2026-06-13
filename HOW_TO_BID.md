# How to Bid on Auctions

This guide explains how to place bids on both **English** and **Dutch** auctions in the auction platform.

## Prerequisites

Before you can bid on an auction, you must:

1. **Have a registered account** - Use the `AuthService.Register` endpoint
2. **Be logged in** - Use the `AuthService.Login` endpoint to get a JWT token
3. **Have verified your email** - Use the `AuthService.Verify` endpoint with your OTP code
4. **Have sufficient balance** - Your user account must have enough available balance to cover the bid amount

## Bidding API

### Endpoint

```
POST /auction_platform.v1.AuctionService/BidAuction
```

### Request Format

```json
{
  "auction_id": "uuid-of-auction",
  "user_id": "uuid-of-bidder",
  "amount": 5000,
  "is_auto_bid": false
}
```

### Request Parameters

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `auction_id` | string | Yes | Valid UUID | The unique identifier of the auction |
| `user_id` | string | Yes | Valid UUID | Your user ID (from JWT token) |
| `amount` | int64 | Yes | > 0 | Bid amount in **cents** (e.g., 5000 = $50.00) |
| `is_auto_bid` | bool | No | - | Reserved for future auto-bidding feature |

### Response Format

```json
{
  "bid": {
    "id": "bid-uuid",
    "auction_id": "auction-uuid",
    "user_id": "user-uuid",
    "amount": 5000,
    "status": "BID_STATUS_ACTIVE",
    "is_auto_bid": false,
    "created_at": "2026-06-13T10:30:00Z",
    "updated_at": "2026-06-13T10:30:00Z"
  },
  "message": "Bid placed successfully",
  "timestamp": "2026-06-13T10:30:00Z"
}
```

## Bidding Rules

### English Auctions

English auctions are **ascending price** auctions where bidders compete by placing increasingly higher bids.

#### Rules:
- ✅ Your bid must be **higher** than the current price
- ✅ Multiple bids are allowed until the auction ends
- ✅ The auction may extend its end time if `extend_on_bid` is enabled
- ✅ When outbid, your held funds are automatically released
- ✅ You can increase your own bid (only the difference is held)
- ⏰ Highest bidder when time expires wins

#### Example Flow:

```
1. Auction starts at $50.00 (starting_price: 5000)
2. User A bids $60.00 (amount: 6000) ✅ Valid
3. User B bids $70.00 (amount: 7000) ✅ Valid - User A's funds released
4. User A bids $80.00 (amount: 8000) ✅ Valid - User B's funds released
5. Auction ends - User A wins at $80.00
```

### Dutch Auctions

Dutch auctions are **descending price** auctions that end immediately when the first bid is accepted.

#### Rules:
- ✅ Your bid must be **equal to or greater than** the current price
- ✅ **First bid wins** - auction ends immediately
- ✅ No time extensions
- ✅ Price drops at regular intervals until someone bids
- 🎯 Accept the current price to win instantly

#### Example Flow:

```
1. Auction starts at $100.00 (starting_price: 10000)
2. Price drops by $5.00 every 60 seconds
3. After 5 minutes: current_price = $75.00
4. User A bids $75.00 (amount: 7500) ✅ Valid
5. Auction ENDS immediately - User A wins at $75.00
```

## Balance Management

The platform uses a **dual-balance system** to ensure bid safety:

### Available Balance vs. Held Balance

- **Available Balance**: Money you can use to place new bids
- **Held Balance**: Money locked for your active bids

When you place a bid:
1. The bid amount is moved from `available_balance` to `held_balance`
2. Your funds are "held" (locked) until you're outbid or the auction ends
3. If outbid, your funds are automatically released back to `available_balance`
4. If you win, the `held_balance` is transferred to the seller

### Insufficient Balance Example

```
User Balance:
- available_balance: $45.00
- held_balance: $0.00

Auction current_price: $50.00

Bid attempt: $60.00 ❌ FAILS
Error: "Insufficient balance"
```

## Step-by-Step: How to Bid

### 1. Find an Active Auction

Use the `GetAuctionsList` endpoint to find auctions:

```json
{
  "status": "AUCTION_STATUS_ACTIVE",
  "type": "AUCTION_TYPE_ENGLISH",
  "page": 1,
  "page_size": 10
}
```

### 2. Check Auction Details

Use `GetAuctionDetailsById` to get full auction information:

```json
{
  "auction_id": "your-auction-uuid"
}
```

Look for:
- `current_price`: The minimum bid you must beat (English) or match (Dutch)
- `end_time`: When the auction closes
- `status`: Must be `AUCTION_STATUS_ACTIVE`
- `type`: `AUCTION_TYPE_ENGLISH` or `AUCTION_TYPE_DUTCH`

### 3. Calculate Your Bid

#### For English Auctions:
```
Your bid > current_price
Example: current_price = $50.00 → bid at least $50.01 (5001 cents)
```

#### For Dutch Auctions:
```
Your bid >= current_price
Example: current_price = $50.00 → bid at least $50.00 (5000 cents)
```

### 4. Check Your Balance

Ensure you have enough available balance:
```
available_balance >= bid_amount
```

### 5. Submit Your Bid

Make the API call:

```bash
curl -X POST https://your-server.com/auction_platform.v1.AuctionService/BidAuction \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "auction_id": "auction-uuid",
    "user_id": "your-user-uuid",
    "amount": 6000,
    "is_auto_bid": false
  }'
```

## Common Errors and Solutions

### "Auction not active"
- The auction has ended, been cancelled, or hasn't started yet
- Check the `status` field - must be `AUCTION_STATUS_ACTIVE`

### "Bid must be higher than current price"
- **English auction**: Your bid is <= current price
- Solution: Bid at least `current_price + 1` cent

### "Bid must match or exceed current price"
- **Dutch auction**: Your bid is < current price
- Solution: Bid at least `current_price`

### "Insufficient balance"
- Your `available_balance` is less than the bid amount
- Solution: Add funds to your account or bid a lower amount

### "Cannot bid on your own auction"
- You are the seller of this auction
- Solution: Sellers cannot bid on their own auctions

### "Auction ended already"
- The auction's `end_time` has passed
- Solution: Find a different active auction

### "New bid must be higher than your current bid"
- You're trying to increase your existing bid but the amount is not higher
- Solution: Bid more than your current active bid amount

## Increasing Your Own Bid

If you already have an active bid on an auction, you can increase it:

1. Your new bid must be **higher** than your current bid
2. Only the **difference** is deducted from your available balance
3. Your old bid is marked as `BID_STATUS_OUTBID`
4. A new bid record is created with status `BID_STATUS_ACTIVE`

### Example:

```
Initial State:
- Your active bid: $60.00
- Available balance: $50.00
- Held balance: $60.00

New bid: $75.00
- Difference: $15.00 ($75 - $60)
- New available balance: $35.00 ($50 - $15)
- New held balance: $75.00
```

## Time Extensions (English Auctions Only)

Some English auctions are configured to extend their end time when bids are placed near the deadline:

- If `extend_on_bid: true` and `extend_minutes: 5`
- Every new bid adds 5 minutes to the `end_time`
- This prevents "sniping" (last-second bidding)
- Check `original_end_time` to see when the auction was originally scheduled to end

## Bid Status Lifecycle

```
BID_STATUS_ACTIVE     → Your bid is currently the highest
BID_STATUS_OUTBID     → Someone bid higher than you
BID_STATUS_WON        → You won the auction
BID_STATUS_LOST       → Auction ended, you didn't win
BID_STATUS_REFUNDED   → Your bid was refunded (auction cancelled)
```

## Best Practices

### For English Auctions:
1. **Bid early** if the auction has time extensions to discourage competition
2. **Monitor closely** near the end time if no extensions are configured
3. **Bid incrementally** to avoid overpaying
4. **Check your balance** before the auction ends to increase your bid if needed

### For Dutch Auctions:
1. **Decide your maximum price** before watching the auction
2. **Wait for the price to drop** to your target
3. **Bid immediately** when the price is acceptable (first bid wins!)
4. **Don't hesitate** - others are watching the same auction

## Security Notes

- All bid amounts are validated server-side
- Balance checks happen within database transactions with row-level locking
- You cannot bid with insufficient funds
- Concurrent bids are handled safely with pessimistic locking
- All bid activity is recorded in the auction history for auditing

## Example Code (Go)

```go
package main

import (
    "context"
    "log"
    
    "connectrpc.com/connect"
    v1 "github.com/KrishnaGrg1/auction_platform/gen/auction_platform/v1"
)

func placeBid(client v1.AuctionServiceClient, auctionID, userID string, amount int64) {
    req := connect.NewRequest(&v1.BidAuctionRequest{
        AuctionId: auctionID,
        UserId:    userID,
        Amount:    amount,
        IsAutoBid: false,
    })
    
    resp, err := client.BidAuction(context.Background(), req)
    if err != nil {
        log.Fatalf("Bid failed: %v", err)
    }
    
    log.Printf("Bid placed successfully! Bid ID: %s", resp.Msg.Bid.Id)
    log.Printf("Status: %s", resp.Msg.Bid.Status)
    log.Printf("Amount: $%.2f", float64(resp.Msg.Bid.Amount)/100)
}
```

## Next Steps

- **Track your bids**: Use `GetUserAuctions` to see your active bids
- **View auction history**: Check the bid history of any auction
- **Monitor balances**: Keep track of your available vs. held balance
- **Set up notifications**: (Coming soon) Get alerts when you're outbid

---

For more information, see:
- [README.md](./README.md) - Platform overview
- [API Documentation](./auction_platform/v1/auction.proto) - Full API specification
