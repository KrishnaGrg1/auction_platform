package store

import (
	"context"
	"log"

	db "github.com/KrishnaGrg1/auction_platform/internal/db/sqlc"
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

func New(pool *pgxpool.Pool, socket *socket.Hub) *Store {
	return &Store{
		Pool:      pool,
		Queries:   db.New(pool),
		socketHub: socket,
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
	log.Println("Connected to Neon database")
	return New(pool, socket.NewHub()), nil
}
