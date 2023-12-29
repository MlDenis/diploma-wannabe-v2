package jobmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/logger"
	"github.com/MlDenis/diploma-wannabe-v2/internal/mocks"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
)

type TestHandler struct {
	*chi.Mux
	Cursor  *db.Cursor
	Manager *Jobmanager
}

func TestJobmanager(t *testing.T) {
	cursor := &db.Cursor{IDBInterface: mocks.NewMock()}
	ctx := context.Background()
	handler := &TestHandler{
		chi.NewMux(),
		cursor,
		NewJobmanager(cursor, "localhost:8081", &ctx),
	}
	ts := httptest.NewServer(handler)
	err := handler.Cursor.SaveUserInfo(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		return
	}

	defer ts.Close()

	defer func(method, url string, body io.Reader) {
		_, err := http.NewRequest(method, url, body)
		if err != nil {
			return
		}
	}(http.MethodGet, "http://localhost:8081/shutdown", nil)

	go func() {
		initMockAccrual("localhost:8081")
	}()
	client := &http.Client{}

	orders := []string{"11111111", "22222222"}
	go handler.Manager.ManageJobs("http://localhost:8081")

	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		return
	}
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/login", buff)
	request.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, request)

	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	cookies := res.Cookies()

	assert.Equal(t, 200, res.StatusCode)

	for _, order := range orders {
		buff := bytes.NewBuffer([]byte{})
		_, _ = buff.Write([]byte(order))
		createOrder := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/orders", buff)
		createOrder.Header.Add("Content-Type", "text/plain")
		createOrder.AddCookie(cookies[0])

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, createOrder)
		res := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(res.Body)

		assert.Equal(t, 202, res.StatusCode)

		createAccrualOrder, _ := http.NewRequest(http.MethodPost, "http://localhost:8081/api/orders/"+order, nil)
		resp, err := client.Do(createAccrualOrder)
		if err != nil {
			logger.ErrorLog.Fatal(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(resp.Body)
		assert.Equal(t, 200, resp.StatusCode)
		err = handler.Manager.AddJob("test", order)
		if err != nil {
			return
		}
	}
	time.Sleep(2 * time.Second)
	request = httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/user/orders", nil)
	request.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	response := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	assert.Equal(t, 200, response.StatusCode)
	result := []*models.Order{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		logger.ErrorLog.Fatal(err)
	}

	assert.Equal(t, "11111111", result[0].Number)
	assert.Equal(t, "PROCESSING", result[0].Status)

	assert.Equal(t, "22222222", result[1].Number)
	assert.Equal(t, "INVALID", result[1].Status)

	time.Sleep(2 * time.Second)

	request = httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/user/orders", nil)
	request.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, request)
	response = w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	assert.Equal(t, 200, response.StatusCode)
	result = []*models.Order{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		logger.ErrorLog.Fatal(err)
	}

	assert.Equal(t, "11111111", result[0].Number)
	assert.Equal(t, "PROCESSED", result[0].Status)
	assert.Equal(t, float64(100), result[0].Accrual)

	assert.Equal(t, "22222222", result[1].Number)
	assert.Equal(t, "INVALID", result[1].Status)
}
