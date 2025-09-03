package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"subscription-aggregator/internal/config"
	"subscription-aggregator/internal/db"
	"subscription-aggregator/internal/models"
)

type Storage struct {
	database *pgx.Conn
	logger   *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Storage, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)

	database, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		logger.Error("Unable to connect to database", "error", err)
		return nil, err
	}

	if err := database.Ping(ctx); err != nil {
		logger.Error("Ping to connect database failed", "error", err)
		database.Close(ctx)
		return nil, err
	}

	logger.Info("Connected to database")

	return &Storage{
		database: database,
		logger:   logger,
	}, nil
}

func (s *Storage) Save(ctx context.Context, sub *models.Subscription) error {
	sql := `INSERT INTO subscriptions (id, service_name, price, user_ID, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := s.database.Exec(
		ctx,
		sql,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	); err != nil {
		s.logger.Error("Unable to save subscription", "error", err)
		return fmt.Errorf("unable to save subscription: %w", err)
	}

	s.logger.Info("Subscription saved successfully", "ID", sub.ID)

	return nil
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	sql := `DELETE FROM subscriptions WHERE id = $1`
	result, err := s.database.Exec(ctx, sql, id)
	if err != nil {
		s.logger.Error("Failed to delete subscription", "error", err)
		return err
	}

	if result.RowsAffected() == 0 {
		s.logger.Error("Failed to find subscription", "error", err, "id", id)
		return db.ErrNotFound
	}

	s.logger.Info("Subscription deleted successfully", "ID", id)
	return nil

}

func (s *Storage) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	sql := `SELECT id, service_name, price, user_ID, start_date, end_date FROM subscriptions WHERE id = $1`

	var sub models.Subscription
	err := s.database.QueryRow(ctx, sql, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error("Failed to find subscription", "error", err, "id", id)
			return nil, db.ErrNotFound
		}

		s.logger.Error("Failed to get subscription", "error", err)
		return nil, fmt.Errorf("failed to get subscription by id: %w", err)
	}

	s.logger.Info("Subscription found successfully", "ID", id)

	return &sub, nil
}

func (s *Storage) Close(ctx context.Context) error {
	if s.database != nil {
		return s.database.Close(ctx)
	}

	return nil
}
