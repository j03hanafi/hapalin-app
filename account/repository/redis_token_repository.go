package repository

import (
	"context"
	"fmt"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type redisTokenRepository struct {
	Redis *redis.Client
}

// NewTokenRepository is a factory for initializing User Repositories
func NewTokenRepository(redis *redis.Client) domain.TokenRepository {
	return &redisTokenRepository{
		Redis: redis,
	}
}

// SetRefreshToken stores a refresh token with an expiry time
func (r redisTokenRepository) SetRefreshToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error {
	l := logger.Get()

	// We'll store userID with token id, so we can scan (non-blocking)
	// over the user's tokens and delete them in case of token leakage
	key := fmt.Sprintf("%s:%s", userID, tokenID)
	if err := r.Redis.Set(ctx, key, 0, expiresIn).Err(); err != nil {
		l.Error("Could not SET refresh token to redis",
			zap.String("userID", userID),
			zap.String("tokenID", tokenID),
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	return nil
}

// DeleteRefreshToken used to delete old  refresh tokens
// Services my access this to revolve tokens
func (r redisTokenRepository) DeleteRefreshToken(ctx context.Context, userID, prevTokenID string) error {
	l := logger.Get()
	key := fmt.Sprintf("%s:%s", userID, prevTokenID)
	if err := r.Redis.Del(ctx, key).Err(); err != nil {
		l.Error("Could not delete refresh token to redis",
			zap.String("userID", userID),
			zap.String("prevTokenID", prevTokenID),
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	return nil
}
