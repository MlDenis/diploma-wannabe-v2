package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (h *BalanceRouter) GetBalance(rw http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session_token")
	sessionToken := cookie.Value
	username, err := h.Cursor.GetUsernameByToken(sessionToken)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	balance, err := h.Cursor.GetUserBalance(username)
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
