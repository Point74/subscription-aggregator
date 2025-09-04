package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
	_ "subscription-aggregator/api/docs"
	"subscription-aggregator/internal/config"
	"subscription-aggregator/internal/db/postgres"
	"subscription-aggregator/internal/handlers"
	"subscription-aggregator/internal/logger"
)

// @title Subscription Aggregator API
// @version 1.0
// @description REST service for aggregating data about users' online subscriptions
// @host localhost:8080
// @BasePath '/'
func main() {
	log := logger.NewLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Error loading config: ", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage, err := postgres.New(ctx, cfg, log)
	if err != nil {
		log.Error("Error creating storage: ", err)
		os.Exit(1)
	}

	defer storage.Close(ctx)

	subscriptionHandler := handlers.NewSubscriptionsHandler(storage, log)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Service start"))
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	router.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", subscriptionHandler.CreateSubscription)
		r.Delete("/{id}", subscriptionHandler.DeleteSubscription)
		r.Get("/{id}", subscriptionHandler.GetSubscriptionByID)
		r.Get("/", subscriptionHandler.ListSubscriptionsByUserID)
		r.Put("/{id}", subscriptionHandler.UpdateSubscription)
		r.Get("/total-cost", subscriptionHandler.SumTotalCostSubscriptions)
	})

	log.Info("Service start on port :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Error("Error starting server: ", err)
		os.Exit(1)
	}
}
