package jobmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/MlDenis/diploma-wannabe-v2/internal/logger"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type MockAccrualHandler struct {
	*chi.Mux
	mu             sync.RWMutex
	OrdersStorage  map[string]*models.AccrualResponse
	ShutdownButton chan struct{}
}

func (mock *MockAccrualHandler) GetAccrual(rw http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")
	response, ok := mock.OrdersStorage[number]
	if !ok {
		rw.WriteHeader(http.StatusNoContent)
	}
	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	encoder.Encode(response)
	rw.Write(buff.Bytes())
}

func (mock *MockAccrualHandler) CreateOrder(rw http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	go func() {
		mock.mu.Lock()
		mock.OrdersStorage[number] = &models.AccrualResponse{
			Order:  number,
			Status: "REGISTERED",
		}
		mock.mu.Unlock()
		time.Sleep(2 * time.Second)
		n, _ := strconv.Atoi(number)
		mock.mu.Lock()
		if n%2 != 0 {
			mock.OrdersStorage[number].Status = "PROCESSING"
		} else {
			mock.OrdersStorage[number].Status = "INVALID"
		}
		mock.mu.Unlock()

		time.Sleep(2 * time.Second)

		mock.mu.Lock()
		mock.OrdersStorage[number].Accrual = 100
		mock.OrdersStorage[number].Status = "PROCESSED"
		mock.mu.Unlock()
	}()
	rw.WriteHeader(http.StatusOK)
}

func NewMockAccrualHandler() *MockAccrualHandler {
	handler := &MockAccrualHandler{
		Mux:            chi.NewMux(),
		OrdersStorage:  make(map[string]*models.AccrualResponse),
		ShutdownButton: make(chan struct{}),
	}
	handler.Get("/api/orders/{number}", handler.GetAccrual)
	handler.Post("/api/orders/{number}", handler.CreateOrder)
	return handler
}

func initMockAccrual(addr string) *http.Server {
	handler := NewMockAccrualHandler()
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	handler.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		server.Shutdown(context.Background())
	})
	err := server.ListenAndServe()
	if err != nil {
		logger.ErrorLog.Printf("Failed to launch mock accrual server:%+v\n", err)
	}
	return server
}

func TestAccrualValidOrder(t *testing.T) {
	defer httptest.NewRequest(http.MethodGet, "http://localhost:8081/shutdown", nil)

	go func() {
		initMockAccrual("localhost:8081")
	}()
	client := &http.Client{}

	createOrder, _ := http.NewRequest(http.MethodPost, "http://localhost:8081/api/orders/1", nil)
	resp, err := client.Do(createOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	getOrder, _ := http.NewRequest(http.MethodGet, "http://localhost:8081/api/orders/1", nil)
	resp, err = client.Do(getOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	result := &models.AccrualResponse{}
	json.NewDecoder(resp.Body).Decode(result)
	assert.Equal(t, &models.AccrualResponse{
		Order:  "1",
		Status: "REGISTERED",
	}, result)

	time.Sleep(2 * time.Second)

	getOrder, _ = http.NewRequest(http.MethodGet, "http://localhost:8081/api/orders/1", nil)
	resp, err = client.Do(getOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	result = &models.AccrualResponse{}
	json.NewDecoder(resp.Body).Decode(result)
	assert.Equal(t, &models.AccrualResponse{
		Order:  "1",
		Status: "PROCESSING",
	}, result)

	time.Sleep(2 * time.Second)

	getOrder, _ = http.NewRequest(http.MethodGet, "http://localhost:8081/api/orders/1", nil)
	resp, err = client.Do(getOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	result = &models.AccrualResponse{}
	json.NewDecoder(resp.Body).Decode(result)
	assert.Equal(t, &models.AccrualResponse{
		Order:   "1",
		Status:  "PROCESSED",
		Accrual: 100,
	}, result)
}

func TestAccrualInvalidOrder(t *testing.T) {
	defer httptest.NewRequest(http.MethodGet, "http://localhost:8081/shutdown", nil)

	go func() {
		initMockAccrual("localhost:8081")
	}()
	client := &http.Client{}

	createOrder, _ := http.NewRequest(http.MethodPost, "http://localhost:8081/api/orders/2", nil)
	resp, err := client.Do(createOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	getOrder, _ := http.NewRequest(http.MethodGet, "http://localhost:8081/api/orders/2", nil)
	resp, err = client.Do(getOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	result := &models.AccrualResponse{}
	json.NewDecoder(resp.Body).Decode(result)
	assert.Equal(t, &models.AccrualResponse{
		Order:  "2",
		Status: "REGISTERED",
	}, result)

	time.Sleep(2 * time.Second)

	getOrder, _ = http.NewRequest(http.MethodGet, "http://localhost:8081/api/orders/2", nil)
	resp, err = client.Do(getOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	result = &models.AccrualResponse{}
	json.NewDecoder(resp.Body).Decode(result)
	assert.Equal(t, &models.AccrualResponse{
		Order:  "2",
		Status: "INVALID",
	}, result)
}

func TestAccrualNoSuchOrder(t *testing.T) {
	defer httptest.NewRequest(http.MethodGet, "http://localhost:8081/shutdown", nil)

	go func() {
		initMockAccrual("localhost:8081")
	}()
	client := &http.Client{}

	getOrder, _ := http.NewRequest(http.MethodGet, "http://localhost:8081/api/orders/3", nil)
	resp, err := client.Do(getOrder)
	if err != nil {
		logger.ErrorLog.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, 204, resp.StatusCode)
}
