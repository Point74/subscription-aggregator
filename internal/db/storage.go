package db

import (
	"context"
	"errors"
	"subscription-aggregator/internal/models"
)

type SubscriptionStorage interface {
	Save(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id string) error
}

var (
	ErrNotFound = errors.New("not found")
)
