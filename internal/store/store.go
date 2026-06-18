package store

import (
	"context"
	"log"

	db "github.com/KrishnaGrg1/auction_platform/internal/db/sqlc"
	"github.com/KrishnaGrg1/auction_platform/internal/helper"
	"github.com/KrishnaGrg1/auction_platform/internal/socket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool      *pgxpool.Pool
	Queries   *db.Queries
	socketHub *socket.Hub
}

func (s *Store) SocketHub() *socket.Hub {
	return s.socketHub
}

// AuctionExists implements socket.AuctionValidator
// socket calls this without knowing about store package
func (s *Store) AuctionExists(ctx context.Context, auctionID string) bool {
	id, err := helper.ParsedStringToUUID(auctionID)
	if err != nil {
		return false
	}
	_, err = s.Queries.GetAuctionByID(ctx, id)
	return err == nil
}

func New(pool *pgxpool.Pool, hub *socket.Hub) *Store {
	return &Store{
		Pool:      pool,
		Queries:   db.New(pool),
		socketHub: hub,
	}
}

func Connect(dbUrl string) (*Store, error) {

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Cannot ping database:", err)
		pool.Close()
		return nil, err
	}

	log.Println("Connected to database")

	// create store without hub first
	s := &Store{
		Pool:    pool,
		Queries: db.New(pool),
	}

	// store implements AuctionValidator
	// pass store to NewHub — no circular import
	s.socketHub = socket.NewHub(s)

	return s, nil
}
