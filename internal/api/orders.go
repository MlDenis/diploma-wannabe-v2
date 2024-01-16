package api

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"

	"github.com/theplant/luhn"
)

func (h *OrderRouter) UploadOrder(rw http.ResponseWriter, r *http.Request) {
	val := r.Header.Get("Content-Type")
	if val != "text/plain" {
		http.Error(rw, "wrong content", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	cookie, _ := r.Cookie("session_token")
	sessionToken := cookie.Value
	username, err := h.Cursor.GetUsernameByToken(sessionToken, h.Logger)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	requestNumber := string(body)

	n, err := strconv.Atoi(requestNumber)
	if err != nil {
		http.Error(rw, "wrong number format", http.StatusUnprocessableEntity)
		return
	}
	if !luhn.Valid(n) {
		http.Error(rw, "wrong number format", http.StatusUnprocessableEntity)
		return
	}

	order, err := GetOrderFromDB(h.Cursor, username, requestNumber, h.Logger)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if order == nil {
		h.Logger.Info("Adding new order for user", zap.String("", username))
		newOrder := &models.Order{
			Number:     requestNumber,
			Username:   username,
			UploadedAt: time.Now(),
			Status:     "NEW",
		}
		err := ValidateOrder(h.Cursor, newOrder)
		if err != nil {
			h.Logger.Error("Validation error for new order, token", zap.String("", sessionToken))
			http.Error(rw, "order was uploaded already by another user", http.StatusConflict)
			return
		}
		err = h.Cursor.SaveOrder(newOrder, h.Logger)
		if err != nil {
			return
		}
		err = h.Manager.AddJob(requestNumber, username)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusAccepted)
		_, err = rw.Write([]byte(`new order created`))
		if err != nil {
			return
		}
		return
	}

	h.Logger.Info(order.Username)
	if order.Username != username {
		h.Logger.Error("Validation error for order, token", zap.String("", sessionToken))
		http.Error(rw, "order was uploaded already by another user", http.StatusConflict)
		return
	}
	h.Logger.Info("request number", zap.String("", requestNumber))

	if order.Number == requestNumber {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte(`order created already`))
		if err != nil {
			return
		}
	}
}

func GetOrderFromDB(cursor *db.Cursor, username string, requestOrder string, l *zap.Logger) (*models.Order, error) {
	order, err := cursor.GetOrder(username, requestOrder, l)
	if order == nil {
		return nil, err
	}
	return order, nil
}

func (h *OrderRouter) GetOrders(rw http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session_token")
	sessionToken := cookie.Value
	username, err := h.Cursor.GetUsernameByToken(sessionToken, h.Logger)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	orders, err := h.Cursor.GetOrders(username, h.Logger)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	if orders == nil {
		rw.WriteHeader(http.StatusNoContent)
		_, err := rw.Write([]byte(`no orders found`))
		if err != nil {
			return
		}
	} else {
		body := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(body)
		err := encoder.Encode(&orders)
		if err != nil {
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write(body.Bytes())
		if err != nil {
			return
		}
	}
}
