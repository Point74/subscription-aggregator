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
	"time"
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

// CreateSubscription creates a new subscription.
// @Summary Create a new subscription
// @Description Creates a new user subscription. Dates must be in "MM-YYYY" format.
// @Accept json
// @Produce json
// @Param subscription body models.SubscriptionRequest true "Subscription data"
// @Success 201 {object} models.SubscriptionRequest "Subscription created successfully"
// @Failure 400 {string} string "Invalid request body or data"
// @Failure 500 {string} string "Could not save subscription"
// @Router /subscriptions [post]
func (h *SubscriptionsHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req models.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updateRequest, err := utils.MapRequest(req, h.log)
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

// DeleteSubscription deletes a subscription by ID.
// @Summary Delete a subscription
// @Description Deletes a user's subscription record by its unique ID.
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid subscription ID"
// @Failure 500 {string} string "Could not delete subscription"
// @Router /subscriptions/{id} [delete]
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

// GetSubscriptionByID get a subscription by ID.
// @Summary Get a subscription by ID
// @Description Get a user's subscription record by its unique ID.
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.Subscription "Subscription found successfully"
// @Failure 400 {string} string "Invalid subscription ID"
// @Failure 500 {string} string "Could not get subscription"
// @Router /subscriptions/{id} [get]
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

// ListSubscriptionsByUserID list of subscriptions for specific user.
// @Summary Get subscriptions by user ID
// @Description Get a list of all subscription records for a specific user.
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {array} models.Subscription "Subscriptions retrieved successfully"
// @Failure 400 {string} string "No user ID"
// @Failure 500 {string} string "Could not get list subscriptions"
// @Router /subscriptions [get]
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

// UpdateSubscription updates an existing subscription.
// @Summary Update an existing subscription
// @Description Update an existing subscription record by its unique ID.
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body models.SubscriptionRequest true "Updated subscription data"
// @Success 200 {object} models.SubscriptionRequest "Subscription updated successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Could not update subscription"
// @Router /subscriptions/{id} [put]
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

	updateRequest, err := utils.MapRequest(req, h.log)
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

// SumTotalCostSubscriptions calculate total cost of a user's subscriptions for a given period and service.
// @Summary Calculate total subscription cost
// @Description Calculates the total cost of a user's subscriptions for a given period, filtered by user ID and service name. The period dates must be in "MM-YYYY" format.
// @Produce json
// @Param user_id query string true "User ID"
// @Param service_name query string true "Service name"
// @Param period_start query string true "Start date of the period (MM-YYYY)"
// @Param period_end query string false "End date of the period (MM-YYYY)"
// @Success 200 {object} map[string]int "Total cost calculated successfully"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 500 {string} string "Could not calculate total cost"
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionsHandler) SumTotalCostSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "no user ID", http.StatusBadRequest)
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		http.Error(w, "no service name", http.StatusBadRequest)
		return
	}

	var periodStart time.Time

	perStart := r.URL.Query().Get("period_start")
	if perStart != "" {
		parseStartDate, err := utils.ParseDate(perStart)
		if err != nil {
			h.log.Error("failed to parse end date", "error", err)
			return
		}

		periodStart = parseStartDate
	}

	var periodEnd time.Time

	perEnd := r.URL.Query().Get("period_end")
	if perEnd != "" {
		parseEndDate, err := utils.ParseDate(perEnd)
		if err != nil {
			h.log.Error("failed to parse end date", "error", err)
			return
		}

		periodEnd = parseEndDate
	}

	reqID := middleware.GetReqID(r.Context())

	result, err := h.storage.SumTotalCost(r.Context(), userID, serviceName, periodStart, periodEnd)
	if err != nil {
		h.log.Error("could not get sum subscriptions", "error", err, "user_id", userID, "request_id", reqID)
		http.Error(w, "could not get sum subscriptions", http.StatusInternalServerError)
		return
	}

	h.log.Info("Successfully get sum subscriptions", "user_id", userID, "request_id", reqID)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&result); err != nil {
		h.log.Error("failed to write response", "error", err, "user_id", userID, "request_id", reqID)
	}
}
