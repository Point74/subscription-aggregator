package handlers

import (
	"encoding/json"
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
