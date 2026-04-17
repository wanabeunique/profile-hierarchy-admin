package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	mw "test-app/backend/middleware"
	"test-app/backend/models"
	"test-app/backend/types"
)

// profileAggregates holds computed values loaded via raw SQL.
type profileAggregates struct {
	ProfileID       uint    `json:"profile_id"`
	TotalAmount     float64 `json:"total_amount"`
	TotalPaid       float64 `json:"total_paid"`
	TotalRemaining  float64 `json:"total_remaining"`
	OrderCount      int     `json:"order_count"`
	SubAccountCount int     `json:"sub_account_count"`
}

const aggregateSQL = `
	SELECT
		p.id AS profile_id,
		COALESCE((SELECT SUM(amount) FROM orders WHERE profile_id = p.id), 0) AS total_amount,
		COALESCE((SELECT SUM(paid) FROM orders WHERE profile_id = p.id), 0) AS total_paid,
		COALESCE((SELECT SUM(amount) FROM orders WHERE profile_id = p.id), 0) -
			COALESCE((SELECT SUM(paid) FROM orders WHERE profile_id = p.id), 0) AS total_remaining,
		COALESCE((SELECT COUNT(*) FROM orders WHERE profile_id = p.id), 0) AS order_count,
		COALESCE((SELECT COUNT(*) FROM profiles WHERE parent_id = p.id), 0) AS sub_account_count
	FROM profiles p WHERE p.id IN ?`

// loadAggregates fetches computed stats for the given profile IDs and applies them.
func (h *ProfileHandler) loadAggregates(profiles []models.Profile) {
	if len(profiles) == 0 {
		return
	}

	ids := make([]uint, len(profiles))
	for i := range profiles {
		ids[i] = profiles[i].ID
	}

	var aggs []profileAggregates
	h.DB.Raw(aggregateSQL, ids).Scan(&aggs)

	aggMap := make(map[uint]*profileAggregates, len(aggs))
	for i := range aggs {
		aggMap[aggs[i].ProfileID] = &aggs[i]
	}

	for i := range profiles {
		if a, ok := aggMap[profiles[i].ID]; ok {
			profiles[i].TotalAmount = a.TotalAmount
			profiles[i].TotalPaid = a.TotalPaid
			profiles[i].TotalRemaining = a.TotalRemaining
			profiles[i].OrderCount = a.OrderCount
			profiles[i].SubAccountCount = a.SubAccountCount
		}
	}
}

// loadParents fetches parent profiles for profiles that have a ParentID.
func (h *ProfileHandler) loadParents(profiles []models.Profile) {
	parentIDs := make([]uint, 0)
	for _, p := range profiles {
		if p.ParentID != nil {
			parentIDs = append(parentIDs, *p.ParentID)
		}
	}
	if len(parentIDs) == 0 {
		return
	}

	var parents []models.Profile
	h.DB.Where("id IN ?", parentIDs).Find(&parents)

	parentMap := make(map[uint]*models.Profile, len(parents))
	for i := range parents {
		parentMap[parents[i].ID] = &parents[i]
	}
	for i := range profiles {
		if profiles[i].ParentID != nil {
			profiles[i].Parent = parentMap[*profiles[i].ParentID]
		}
	}
}

type ProfileHandler struct {
	DB *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{DB: db}
}

func (h *ProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := mw.GetClaims(r.Context())

	query := h.DB.Model(&models.Profile{})

	if t := r.URL.Query().Get("type"); t != "" {
		query = query.Where("profiles.type = ?", t)
	}

	if pid := r.URL.Query().Get("parent_id"); pid != "" {
		query = query.Where("profiles.parent_id = ?", pid)
	}

	if claims.Role != types.RoleAdmin {
		accessibleIDs := h.getAccessibleProfileIDs(claims)
		query = query.Where("profiles.id IN ?", accessibleIDs)
	}

	var profiles []models.Profile
	if err := query.Find(&profiles).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch profiles")
		return
	}

	h.loadAggregates(profiles)
	h.loadParents(profiles)

	writeJSON(w, http.StatusOK, profiles)
}

func (h *ProfileHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := mw.GetClaims(r.Context())

	req, err := decodeJSON[types.CreateProfileRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" || req.Type == "" {
		writeError(w, http.StatusBadRequest, "name, email, password, and type are required")
		return
	}

	if !isValidProfileType(req.Type) {
		writeError(w, http.StatusBadRequest, "type must be plus, standard, or min")
		return
	}

	if err := h.validateHierarchy(req.Type, req.ParentID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if claims.Role != types.RoleAdmin {
		if err := validateProfileCreationAccess(claims, req.Type, req.ParentID); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	profile := models.Profile{
		Type:            req.Type,
		ParentID:        req.ParentID,
		Name:            req.Name,
		Email:           req.Email,
		PasswordHash:    string(hash),
		Commission:      req.Commission,
		PaymentDeadline: req.PaymentDeadline,
	}

	if err := h.DB.Create(&profile).Error; err != nil {
		if isDuplicateKey(err) {
			writeError(w, http.StatusConflict, "профиль с таким email уже существует")
			return
		}
		writeError(w, http.StatusInternalServerError, "не удалось создать профиль")
		return
	}

	writeJSON(w, http.StatusCreated, profile)
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	claims := mw.GetClaims(r.Context())
	if !mw.CheckProfileAccess(h.DB, claims, id) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var profile models.Profile
	if err := h.DB.First(&profile, id).Error; err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch profile")
		return
	}

	// Load aggregates
	profiles := []models.Profile{profile}
	h.loadAggregates(profiles)
	profile = profiles[0]

	// Load relations
	if profile.ParentID != nil {
		var parent models.Profile
		h.DB.First(&parent, *profile.ParentID)
		profile.Parent = &parent
	}
	h.DB.Where("parent_id = ?", id).Find(&profile.Children)
	h.DB.Where("profile_id = ?", id).Order("created_at DESC").Find(&profile.Orders)

	writeJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	claims := mw.GetClaims(r.Context())
	if !mw.CheckProfileAccess(h.DB, claims, id) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	var profile models.Profile
	if err := h.DB.First(&profile, id).Error; err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch profile")
		return
	}

	req, err := decodeJSON[types.UpdateProfileRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Commission != nil {
		updates["commission"] = *req.Commission
	}
	if req.PaymentDeadline != nil {
		updates["payment_deadline"] = *req.PaymentDeadline
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		updates["password_hash"] = string(hash)
	}

	if len(updates) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := h.DB.Model(&profile).Updates(updates).Error; err != nil {
		if isDuplicateKey(err) {
			writeError(w, http.StatusConflict, "профиль с таким email уже существует")
			return
		}
		writeError(w, http.StatusInternalServerError, "не удалось обновить профиль")
		return
	}

	// Re-fetch with aggregates
	h.DB.First(&profile, profile.ID)
	profiles := []models.Profile{profile}
	h.loadAggregates(profiles)
	profile = profiles[0]

	writeJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	claims := mw.GetClaims(r.Context())
	if claims.Role != types.RoleAdmin {
		if !mw.CheckProfileAccess(h.DB, claims, id) || claims.ProfileID == id {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var childIDs []uint
		tx.Model(&models.Profile{}).Where("parent_id = ?", id).Pluck("id", &childIDs)

		var grandchildIDs []uint
		if len(childIDs) > 0 {
			tx.Model(&models.Profile{}).Where("parent_id IN ?", childIDs).Pluck("id", &grandchildIDs)
		}

		if len(grandchildIDs) > 0 {
			if err := tx.Where("profile_id IN ?", grandchildIDs).Delete(&models.Order{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", grandchildIDs).Delete(&models.Profile{}).Error; err != nil {
				return err
			}
		}

		if len(childIDs) > 0 {
			if err := tx.Where("profile_id IN ?", childIDs).Delete(&models.Order{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", childIDs).Delete(&models.Profile{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("profile_id = ?", id).Delete(&models.Order{}).Error; err != nil {
			return err
		}

		return tx.Delete(&models.Profile{}, id).Error
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete profile: "+err.Error())
		return
	}

	writeMessage(w, http.StatusOK, "profile deleted")
}

// --- Private helpers ---

func isValidProfileType(t string) bool {
	return t == types.ProfileTypePlus || t == types.ProfileTypeStandard || t == types.ProfileTypeMin
}

func (h *ProfileHandler) validateHierarchy(profileType string, parentID *uint) error {
	switch profileType {
	case types.ProfileTypePlus:
		if parentID != nil {
			return errors.New("plus profiles cannot have a parent")
		}
	case types.ProfileTypeStandard:
		if parentID != nil {
			var parent models.Profile
			if err := h.DB.First(&parent, *parentID).Error; err != nil {
				return errors.New("parent profile not found")
			}
			if parent.Type != types.ProfileTypePlus {
				return errors.New("standard profiles can only be children of plus profiles")
			}
		}
	case types.ProfileTypeMin:
		if parentID != nil {
			var parent models.Profile
			if err := h.DB.First(&parent, *parentID).Error; err != nil {
				return errors.New("parent profile not found")
			}
			if parent.Type != types.ProfileTypeStandard {
				return errors.New("min profiles can only be children of standard profiles")
			}
		}
	}
	return nil
}

func validateProfileCreationAccess(claims *types.JWTClaims, newType string, parentID *uint) error {
	allowedChild := map[string]string{
		types.ProfileTypePlus:     types.ProfileTypeStandard,
		types.ProfileTypeStandard: types.ProfileTypeMin,
	}

	expected, canCreate := allowedChild[claims.ProfileType]
	if !canCreate {
		return errors.New("you do not have permission to create profiles")
	}
	if newType != expected {
		return fmt.Errorf("%s profiles can only create %s sub-profiles", claims.ProfileType, expected)
	}
	if parentID == nil || *parentID != claims.ProfileID {
		return errors.New("you can only create profiles under your own account")
	}
	return nil
}

func (h *ProfileHandler) getAccessibleProfileIDs(claims *types.JWTClaims) []uint {
	ids := []uint{claims.ProfileID}

	switch claims.ProfileType {
	case types.ProfileTypePlus:
		var childIDs []uint
		h.DB.Model(&models.Profile{}).Where("parent_id = ?", claims.ProfileID).Pluck("id", &childIDs)
		ids = append(ids, childIDs...)
		if len(childIDs) > 0 {
			var grandchildIDs []uint
			h.DB.Model(&models.Profile{}).Where("parent_id IN ?", childIDs).Pluck("id", &grandchildIDs)
			ids = append(ids, grandchildIDs...)
		}
	case types.ProfileTypeStandard:
		var childIDs []uint
		h.DB.Model(&models.Profile{}).Where("parent_id = ?", claims.ProfileID).Pluck("id", &childIDs)
		ids = append(ids, childIDs...)
	}

	return ids
}
