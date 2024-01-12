package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
	"github.com/google/uuid"
)

func (h *UserRouter) Login(rw http.ResponseWriter, r *http.Request) {
	userInput := &models.UserInfo{}
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateUserInfo(userInput); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	dbData, err := h.Cursor.GetUserInfo(userInput, h.Logger)

	if err != nil {
		http.Error(rw, "wrong password/username", http.StatusUnauthorized)
		return
	}
	if err := ValidateLogin(userInput, dbData); err != nil {
		http.Error(rw, "wrong password/username", http.StatusUnauthorized)
		return
	}
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(600 * time.Second)

	_ = h.Cursor.SaveSession(sessionToken, &models.Session{
		Username:  userInput.Username,
		ExpiresAt: expiresAt,
		Token:     sessionToken,
	}, h.Logger)

	http.SetCookie(rw, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
	rw.WriteHeader(http.StatusOK)

	_, _ = rw.Write([]byte(`success`))
}
