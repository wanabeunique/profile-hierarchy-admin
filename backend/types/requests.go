package types

import "time"

// LoginRequest is the payload for POST /api/auth/login.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// CreateProfileRequest is the payload for POST /api/profiles.
type CreateProfileRequest struct {
	Type            string     `json:"type"`
	ParentID        *uint      `json:"parentId"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Password        string     `json:"password"`
	Commission      float64    `json:"commission"`
	PaymentDeadline *time.Time `json:"paymentDeadline"`
}

// UpdateProfileRequest is the payload for PUT /api/profiles/:id.
type UpdateProfileRequest struct {
	Name            *string    `json:"name"`
	Email           *string    `json:"email"`
	Password        *string    `json:"password"`
	Commission      *float64   `json:"commission"`
	PaymentDeadline *time.Time `json:"paymentDeadline"`
}

// CreateOrderRequest is the payload for POST /api/profiles/:id/orders.
type CreateOrderRequest struct {
	Number string  `json:"number"`
	Amount float64 `json:"amount"`
	Paid   float64 `json:"paid"`
}

// PayOrderRequest is the payload for PATCH /api/orders/:id/pay.
type PayOrderRequest struct {
	Amount float64 `json:"amount"`
}
