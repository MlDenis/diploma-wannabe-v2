package api

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

func (h *BalanceRouter) GetBalance(rw http.ResponseWriter, r *http.Request, l *zap.Logger) {
	cookie, _ := r.Cookie("session_token")
	sessionToken := cookie.Value
	username, err := h.Cursor.GetUsernameByToken(sessionToken, l)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	balance, err := h.Cursor.GetUserBalance(username, l)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(&balance)
	if err != nil {
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	_, err = rw.Write(buff.Bytes())
	if err != nil {
		return
	}
}
