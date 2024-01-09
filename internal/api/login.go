package api

import (
	"encoding/json"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
	"net/http"
	"time"

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
	dbData, err := h.Cursor.GetUserInfo(userInput)

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

	err = h.Cursor.SaveSession(sessionToken, &models.Session{
		Username:  userInput.Username,
		ExpiresAt: expiresAt,
		Token:     sessionToken,
	})
	if err != nil {
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
	rw.WriteHeader(http.StatusOK)

	_, err = rw.Write([]byte(`success`))
	if err != nil {
		return
	}
}
