package types

import "github.com/golang-jwt/jwt/v5"

// Role constants.
const (
	RoleAdmin   = "admin"
	RoleProfile = "profile"
)

// ProfileType constants.
const (
	ProfileTypePlus     = "plus"
	ProfileTypeStandard = "standard"
	ProfileTypeMin      = "min"
)

// JWTClaims are the typed claims stored in a JWT token.
type JWTClaims struct {
	jwt.RegisteredClaims
	Role        string `json:"role"`
	ProfileID   uint   `json:"profile_id,omitempty"`
	ProfileType string `json:"profile_type,omitempty"`
}
