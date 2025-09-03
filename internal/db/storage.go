package db

import (
	"context"
	"subscription-aggregator/internal/models"
)

type SubscriptionStorage interface {
	Save(ctx context.Context, sub *models.Subscription) error
}
