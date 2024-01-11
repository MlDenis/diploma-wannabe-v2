package api

import (
	"encoding/json"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (h *UserRouter) RegisterUser(rw http.ResponseWriter, r *http.Request, l *zap.Logger) {
	userInput := &models.UserInfo{}
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateUserInfo(userInput); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Cursor.SaveUserInfo(userInput, l); err != nil {
		http.Error(rw, "user already exists", http.StatusConflict)
		return
	}
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(600 * time.Second)

	err := h.Cursor.SaveSession(sessionToken, &models.Session{
		Username:  userInput.Username,
		ExpiresAt: expiresAt,
		Token:     sessionToken,
	}, l)
	if err != nil {
		return
	}
	_, err = h.Cursor.SaveUserBalance(userInput.Username, &models.Balance{
		User:      userInput.Username,
		Current:   0.0,
		Withdrawn: 0.0,
	}, l)
	if err != nil {
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write([]byte(`user created successfully`))
	if err != nil {
		return
	}
}
