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
	"time"
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

func (s *Storage) List(ctx context.Context, userID string) ([]*models.Subscription, error) {
	sql := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE user_id = $1`

	rows, err := s.database.Query(ctx, sql, userID)
	if err != nil {
		s.logger.Error("Failed to list subscriptions", "error", err, "user_id", userID)
		return nil, err
	}

	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		var sub models.Subscription
		scan := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
		)

		if err := scan; err != nil {
			s.logger.Error("Failed to scan subscription row", "error", err, "user_id", userID)
			return nil, err
		}

		subs = append(subs, &sub)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error rows iterations", "error", err, "user_id", userID)
		return nil, err
	}

	s.logger.Info("Subscriptions listed successfully", "user_id", userID)

	return subs, nil
}

func (s *Storage) Update(ctx context.Context, sub *models.Subscription) error {
	sql := `UPDATE subscriptions SET service_name = $1, price = $2, user_ID = $3, start_date = $4, end_date = $5 WHERE id = $6`

	result, err := s.database.Exec(
		ctx,
		sql,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	)
	if err != nil {
		s.logger.Error("Failed to update subscription", "error", err)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		s.logger.Error("Failed to find subscription for update", "error", err, "id", sub.ID)
		return db.ErrNotFound
	}

	s.logger.Info("Subscription updated successfully", "ID", sub.ID)

	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	if s.database != nil {
		return s.database.Close(ctx)
	}

	return nil
}

func (s *Storage) SumTotalCost(ctx context.Context, userID string, serviceName string, periodStart time.Time, periodEnd time.Time) (int, error) {
	sql := `
      WITH subs_in_period AS (
       SELECT 
         s.id, 
         s.user_id, 
         s.service_name, 
         s.price, 
         s.start_date, 
         s.end_date, 
         GREATEST(s.start_date, $1::date) AS actual_start, 
         LEAST(COALESCE(s.end_date, $2::date), $2::date) AS actual_end
       FROM subscriptions s
       WHERE s.user_id = $3
         AND s.service_name = $4
         AND (s.end_date IS NULL OR s.end_date > $1::date)
         AND s.start_date < $2::date
      )
      SELECT COALESCE(SUM(
        price * GREATEST(0, (
          EXTRACT(YEAR FROM date_trunc('month', actual_end - interval '1 day')) * 12 + 
          EXTRACT(MONTH FROM date_trunc('month', actual_end - interval '1 day'))
        ) - (
          EXTRACT(YEAR FROM date_trunc('month', actual_start)) * 12 + 
          EXTRACT(MONTH FROM date_trunc('month', actual_start))
        ) + 1)
      ), 0) AS total_cost
      FROM subs_in_period
      WHERE actual_start < actual_end;
    `

	var totalCost int64
	row := s.database.QueryRow(ctx, sql, periodStart, periodEnd, userID, serviceName)
	if err := row.Scan(&totalCost); err != nil {
		return 0, err
	}

	return int(totalCost), nil
}
