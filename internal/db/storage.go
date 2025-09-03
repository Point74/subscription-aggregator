package db

import (
	"context"
	"errors"
	"subscription-aggregator/internal/models"
)

type SubscriptionStorage interface {
	Save(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	List(ctx context.Context, userID string) ([]*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
}

var (
	ErrNotFound = errors.New("not found")
)
