package handlers

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	mw "test-app/backend/middleware"
	"test-app/backend/models"
	"test-app/backend/types"
)

type OrderHandler struct {
	DB *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{DB: db}
}

func (h *OrderHandler) ListByProfile(w http.ResponseWriter, r *http.Request) {
	profileID, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	claims := mw.GetClaims(r.Context())
	if !mw.CheckProfileAccess(h.DB, claims, profileID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var orders []models.Order
	if err := h.DB.Where("profile_id = ?", profileID).Order("created_at DESC").Find(&orders).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	profileID, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	claims := mw.GetClaims(r.Context())
	if !mw.CheckProfileAccess(h.DB, claims, profileID) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var profile models.Profile
	if err := h.DB.First(&profile, profileID).Error; err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to verify profile")
		return
	}

	req, err := decodeJSON[types.CreateOrderRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Number == "" || req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "number and positive amount are required")
		return
	}
	if req.Paid < 0 {
		writeError(w, http.StatusBadRequest, "paid cannot be negative")
		return
	}
	if req.Paid > req.Amount {
		writeError(w, http.StatusBadRequest, "paid cannot exceed amount")
		return
	}

	order := models.Order{
		ProfileID: profileID,
		Number:    req.Number,
		Amount:    req.Amount,
		Paid:      req.Paid,
	}

	if err := h.DB.Create(&order).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create order: "+err.Error())
		return
	}

	order.ComputeFields()

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.findOrderWithAccess(w, r, orderID)
	if err != nil {
		return
	}

	req, err := decodeJSON[types.CreateOrderRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]any)
	if req.Number != "" {
		updates["number"] = req.Number
	}
	if req.Amount > 0 {
		updates["amount"] = req.Amount
	}
	if req.Paid >= 0 {
		updates["paid"] = req.Paid
	}

	if err := h.DB.Model(&order).Updates(updates).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update order: "+err.Error())
		return
	}

	h.DB.First(&order, order.ID)

	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.findOrderWithAccess(w, r, orderID)
	if err != nil {
		return
	}

	if err := h.DB.Delete(&order).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete order")
		return
	}

	writeMessage(w, http.StatusOK, "order deleted")
}

func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.findOrderWithAccess(w, r, orderID)
	if err != nil {
		return
	}

	req, err := decodeJSON[types.PayOrderRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "payment amount must be positive")
		return
	}

	newPaid := order.Paid + req.Amount
	if newPaid > order.Amount {
		writeError(w, http.StatusBadRequest, "payment would exceed order amount")
		return
	}

	if err := h.DB.Model(&order).Update("paid", newPaid).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process payment")
		return
	}

	h.DB.First(&order, order.ID)

	writeJSON(w, http.StatusOK, order)
}

// findOrderWithAccess loads an order and checks access. On error it writes
// the HTTP response and returns a non-nil error so the caller can return early.
func (h *OrderHandler) findOrderWithAccess(w http.ResponseWriter, r *http.Request, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := h.DB.First(&order, orderID).Error; err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "order not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch order")
		}
		return nil, err
	}

	claims := mw.GetClaims(r.Context())
	if !mw.CheckProfileAccess(h.DB, claims, order.ProfileID) {
		writeError(w, http.StatusForbidden, "access denied")
		return nil, errors.New("access denied")
	}

	return &order, nil
}
