package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"test-app/backend/models"
	"test-app/backend/types"
)

type contextKey string

const contextClaims contextKey = "claims"

// JWTAuth is a middleware that validates the JWT token and stores typed claims in context.
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims := &types.JWTClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts the typed JWT claims from context.
func GetClaims(ctx context.Context) *types.JWTClaims {
	claims, _ := ctx.Value(contextClaims).(*types.JWTClaims)
	return claims
}

// IsAdmin returns true if the context user has the admin role.
func IsAdmin(ctx context.Context) bool {
	c := GetClaims(ctx)
	return c != nil && c.Role == types.RoleAdmin
}

// CheckProfileAccess checks whether the requesting profile can access the target profile.
// Rules:
//   - A profile can always access itself.
//   - "plus" can access its direct children (standard) and their children (min).
//   - "standard" can access its direct children (min).
//   - "min" can only access itself.
func CheckProfileAccess(db *gorm.DB, claims *types.JWTClaims, targetProfileID uint) bool {
	if claims.Role == types.RoleAdmin {
		return true
	}

	if claims.ProfileID == targetProfileID {
		return true
	}

	switch claims.ProfileType {
	case types.ProfileTypePlus:
		var count int64
		db.Model(&models.Profile{}).
			Where("id = ? AND parent_id = ?", targetProfileID, claims.ProfileID).
			Count(&count)
		if count > 0 {
			return true
		}
		db.Model(&models.Profile{}).
			Where("id = ? AND parent_id IN (?)",
				targetProfileID,
				db.Model(&models.Profile{}).Select("id").Where("parent_id = ?", claims.ProfileID),
			).Count(&count)
		return count > 0

	case types.ProfileTypeStandard:
		var count int64
		db.Model(&models.Profile{}).
			Where("id = ? AND parent_id = ?", targetProfileID, claims.ProfileID).
			Count(&count)
		return count > 0

	default:
		return false
	}
}
