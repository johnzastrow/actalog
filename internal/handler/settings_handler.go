package handler

import (
	"encoding/json"
	"net/http"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// SettingsHandler handles user settings endpoints
type SettingsHandler struct {
	settingsService *service.UserSettingsService
	logger          *logger.Logger
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService *service.UserSettingsService, logger *logger.Logger) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		logger:          logger,
	}
}

// GetSettings retrieves user settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=get_settings user_id=%d", userID)
	}

	settings, err := h.settingsService.GetSettings(userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_settings outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve settings")
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

// UpdateSettings updates user settings
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req domain.UserSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=update_settings_attempt user_id=%d", userID)
	}

	settings, err := h.settingsService.UpdateSettings(userID, &req)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=update_settings outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=update_settings outcome=success user_id=%d", userID)
	}

	respondJSON(w, http.StatusOK, settings)
}
