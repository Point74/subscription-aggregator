package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"subscription-aggregator/internal/db/postgres"
	"subscription-aggregator/internal/models"
	"subscription-aggregator/internal/utils"
)

type SubscriptionsHandler struct {
	storage *postgres.Storage
	log     *slog.Logger
}

func NewSubscriptionsHandler(storage *postgres.Storage, log *slog.Logger) *SubscriptionsHandler {
	return &SubscriptionsHandler{
		storage: storage,
		log:     log,
	}
}

func (h *SubscriptionsHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req models.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updateRequest, err := utils.MapRequest(req, h.log, r)
	if err != nil {
		http.Error(w, "invalid request data", http.StatusBadRequest)
		return
	}

	reqID := middleware.GetReqID(r.Context())

	if err := h.storage.Save(r.Context(), updateRequest); err != nil {
		h.log.Error("could not save subscription", "error", err, "subscription_id", updateRequest.ID, "request_id", reqID)
		http.Error(w, "could not save subscription", http.StatusInternalServerError)
		return
	}

	h.log.Info("Successfully saved subscription", "subscription_id", updateRequest.ID, "request_id", reqID)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(&req); err != nil {
		h.log.Error("failed to write response", "error", err, "subscription_id", updateRequest.ID, "request_id", reqID)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionsHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	subID := chi.URLParam(r, "id")
	if subID == "" {
		http.Error(w, "no subscription ID", http.StatusBadRequest)
		return
	}

	reqID := middleware.GetReqID(r.Context())

	if err := h.storage.Delete(r.Context(), subID); err != nil {
		h.log.Error("could not delete subscription", "error", err, "subscription_id", subID, "request_id", reqID)
		http.Error(w, "could not delete subscription", http.StatusInternalServerError)
		return
	}

	h.log.Info("Successfully deleted subscription", "subscription_id", subID, "request_id", reqID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionsHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	subID := chi.URLParam(r, "id")
	if subID == "" {
		http.Error(w, "no subscription ID", http.StatusBadRequest)
		return
	}

	reqID := middleware.GetReqID(r.Context())

	result, err := h.storage.GetByID(r.Context(), subID)
	if err != nil {
		h.log.Error("could not get subscription", "error", err, "subscription_id", subID, "request_id", reqID)
		http.Error(w, "could not get subscription", http.StatusInternalServerError)
		return
	}

	h.log.Info("Successfully get subscription", "subscription_id", subID, "request_id", reqID)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&result); err != nil {
		h.log.Error("failed to write response", "error", err, "subscription_id", subID, "request_id", reqID)
	}
}

func (h *SubscriptionsHandler) ListSubscriptionsByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "no user ID", http.StatusBadRequest)
		return
	}

	reqID := middleware.GetReqID(r.Context())

	result, err := h.storage.List(r.Context(), userID)
	if err != nil {
		h.log.Error("could not get list subscriptions", "error", err, "user_id", userID, "request_id", reqID)
		http.Error(w, "could not get list subscriptions", http.StatusInternalServerError)
		return
	}

	h.log.Info("Successfully get list subscriptions", "user_id", userID, "request_id", reqID)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&result); err != nil {
		h.log.Error("failed to write response", "error", err, "user_id", userID, "request_id", reqID)
	}
}

func (h *SubscriptionsHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	subID := chi.URLParam(r, "id")
	if subID == "" {
		http.Error(w, "no subscription ID", http.StatusBadRequest)
		return
	}

	var req models.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updateRequest, err := utils.MapRequest(req, h.log, r)
	if err != nil {
		http.Error(w, "invalid request data", http.StatusBadRequest)
		return
	}

	updateRequest.ID = subID

	reqID := middleware.GetReqID(r.Context())

	if err := h.storage.Update(r.Context(), updateRequest); err != nil {
		h.log.Error("could not update subscription", "error", err, "subscription_id", subID, "request_id", reqID)
		http.Error(w, "could not update subscription", http.StatusInternalServerError)
		return
	}

	h.log.Info("Successfully update subscription", "subscription_id", subID, "request_id", reqID)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&req); err != nil {
		h.log.Error("failed to write response", "error", err, "subscription_id", subID, "request_id", reqID)
	}
}
