package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func New(dsn, redisAddr string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	r := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := r.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Storage{DB: db, Redis: r}, nil
}

type WalletState struct {
	WalletID string          `json:"wallet_id"`
	Vector   json.RawMessage `json:"vector"`
	HeadHash string          `json:"head_hash"`
	Updated  time.Time       `json:"updated_at"`
}
