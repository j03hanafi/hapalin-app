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

	result := r.Redis.Del(ctx, key)

	if err := result.Err(); err != nil {
		l.Error("Could not delete refresh token to redis",
			zap.String("userID", userID),
			zap.String("prevTokenID", prevTokenID),
			zap.Error(err),
		)
		return apperrors.NewInternal()
	}

	// Val returns count of deleted keys
	// If no key was deleted, the refresh token is invalid
	if result.Val() < 1 {
		l.Error("Refresh token does not exist in redis",
			zap.String("userID", userID),
			zap.String("prevTokenID", prevTokenID),
		)
		return apperrors.NewAuthorization("Invalid refresh token")
	}

	return nil
}

// DeleteUserRefreshTokens looks for all tokens beginning with the userID
// and scans to delete them in a non-blocking way
func (r redisTokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	l := logger.Get()

	pattern := fmt.Sprintf("%s*", userID)

	iter := r.Redis.Scan(ctx, 0, pattern, 5).Iterator()
	failCount := 0

	for iter.Next(ctx) {
		if err := r.Redis.Del(ctx, iter.Val()).Err(); err != nil {
			l.Error("Failed to delete refresh token",
				zap.String("tokenID", iter.Val()),
				zap.Error(err),
			)
			failCount++
		}
	}

	// check last value
	if err := iter.Err(); err != nil {
		l.Error("Failed to delete refresh token",
			zap.String("tokenID", iter.Val()),
			zap.Error(err),
		)
	}

	if failCount > 0 {
		return apperrors.NewInternal()
	}

	return nil
}
