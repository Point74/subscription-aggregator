package utils

import (
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"log/slog"
	"subscription-aggregator/internal/models"
	"time"
)

func MapRequest(req models.SubscriptionRequest, log *slog.Logger) (*models.Subscription, error) {
	startDate, err := ParseDate(req.StartDate)
	if err != nil {
		log.Warn("failed to parse start date", "error", err)
		return nil, err
	}

	var endDate *time.Time

	if req.EndDate != "" {
		parseEndDate, err := ParseDate(req.EndDate)
		if err != nil {
			log.Warn("failed to parse end date", "error", err)
			return nil, err
		}

		endDate = &parseEndDate
	}

	sub := &models.Subscription{}
	if err := copier.Copy(&sub, &req); err != nil {
		log.Warn("failed to copy data to subscription", "err", err)
		return nil, err
	}

	sub.ID = uuid.New().String()
	sub.StartDate = startDate
	sub.EndDate = endDate

	return sub, nil
}
