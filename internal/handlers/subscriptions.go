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

func (h *SubscriptionsHandler) CreateSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updateRequest, err := utils.MapRequest(req, h.log, r)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.storage.Save(r.Context(), updateRequest); err != nil {
		h.log.Error("could not save subscription", "error", err)
		http.Error(w, "could not save subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(&req); err != nil {
		h.log.Error("failed to write response", "error", err, "request_id", updateRequest.ID)
	}
}

func (h *SubscriptionsHandler) DeleteSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
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
