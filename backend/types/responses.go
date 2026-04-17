package types

// ErrorResponse is a standard error JSON payload.
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse is a standard success message JSON payload.
type MessageResponse struct {
	Message string `json:"message"`
}

// LoginResponse is returned after successful authentication.
type LoginResponse struct {
	Token string `json:"token"`
}
