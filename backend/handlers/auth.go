package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"test-app/backend/models"
	"test-app/backend/types"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: jwtSecret}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[types.LoginRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "login and password are required")
		return
	}

	// Try admin user first
	var admin models.AdminUser
	if err := h.DB.Where("username = ?", req.Login).First(&admin).Error; err == nil {
		if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) == nil {
			token, err := h.generateToken(types.RoleAdmin, 0, "")
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate token")
				return
			}
			writeJSON(w, http.StatusOK, types.LoginResponse{Token: token})
			return
		}
	}

	// Try profile by email or name
	var profile models.Profile
	if err := h.DB.Where("email = ? OR name = ?", req.Login, req.Login).First(&profile).Error; err == nil {
		if bcrypt.CompareHashAndPassword([]byte(profile.PasswordHash), []byte(req.Password)) == nil {
			token, err := h.generateToken(types.RoleProfile, profile.ID, profile.Type)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to generate token")
				return
			}
			writeJSON(w, http.StatusOK, types.LoginResponse{Token: token})
			return
		}
	}

	writeError(w, http.StatusUnauthorized, "invalid credentials")
}

func (h *AuthHandler) generateToken(role string, profileID uint, profileType string) (string, error) {
	claims := types.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		Role:        role,
		ProfileID:   profileID,
		ProfileType: profileType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString([]byte(h.JWTSecret))
}
