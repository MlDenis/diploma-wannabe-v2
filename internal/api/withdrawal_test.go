package api

import (
	"bytes"
	"encoding/json"
	"github.com/MlDenis/diploma-wannabe-v2/internal/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawal(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type arguments struct {
		url     string
		payload *models.WithdrawalPost
	}
	tests := []struct {
		name string
		want want
		args arguments
	}{
		{
			name: "Test Positive withdrawal",
			want: want{
				code:     200,
				response: `success`,
			},
			args: arguments{
				url: "http://localhost:8080/api/user/balance/withdraw",
				payload: &models.WithdrawalPost{
					Order: "2377225624",
					Sum:   751,
				},
			},
		},
		{
			name: "Test Negative withdrawal - not enough money",
			want: want{
				code:     402,
				response: "not enough money\n",
			},
			args: arguments{
				url: "http://localhost:8080/api/user/balance/withdraw",
				payload: &models.WithdrawalPost{
					Order: "2377225624",
					Sum:   751,
				},
			},
		},
		{
			name: "Test Negative withdrawal - wrong order number",
			want: want{
				code:     422,
				response: "invalid order number\n",
			},
			args: arguments{
				url: "http://localhost:8080/api/user/balance/withdraw",
				payload: &models.WithdrawalPost{
					Order: "111",
					Sum:   3,
				},
			},
		},
	}
	handler := &Handler{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	ur := &UserRouter{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	br := &BalanceRouter{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	handler.Post("/api/user/login", ur.Login)
	handler.Post("/api/user/balance/withdraw", br.WithdrawMoney)
	ts := httptest.NewServer(handler)
	handler.Cursor.SaveUserInfo(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	handler.Cursor.SaveOrder(
		&models.Order{
			Number:     "2377225624",
			UploadedAt: time.Now(),
		},
	)
	handler.Cursor.UpdateUserBalance(
		"test", &models.Balance{
			User:      "test",
			Current:   752,
			Withdrawn: 0,
		},
	)

	defer ts.Close()

	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	encoder.Encode(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/login", buff)
	request.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()

	cookies := res.Cookies()

	assert.Equal(t, res.StatusCode, 200)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := bytes.NewBuffer([]byte{})
			encoder := json.NewEncoder(buff)
			encoder.Encode(tt.args.payload)
			request := httptest.NewRequest(http.MethodPost, tt.args.url, buff)
			request.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			request.AddCookie(cookies[0])
			handler.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want.response, string(resBody))
		})
	}
}

func TestGetWithdrawal(t *testing.T) {
	layout := "2006-01-02T15:04:05Z07:00"
	parseTime := func(layout string, toParse string) time.Time {
		parsed, _ := time.Parse(layout, toParse)
		return parsed
	}
	mockWithdrawals := []*models.Withdrawal{
		{
			Order:       "2377225624",
			Sum:         500,
			ProcessedAt: parseTime(layout, "2020-12-09T16:09:57+03:00"),
		},
		{
			Order:       "1111111111",
			Sum:         322,
			ProcessedAt: parseTime(layout, "2020-12-09T16:09:57+03:00"),
		},
	}
	type want struct {
		code     int
		response []*models.Withdrawal
	}
	type arguments struct {
		url string
	}
	tests := []struct {
		name string
		want want
		args arguments
	}{
		{
			name: "Test Positive GET withdrawals",
			want: want{
				code:     200,
				response: mockWithdrawals,
			},
			args: arguments{
				url: "http://localhost:8080/api/user/withdrawals",
			},
		},
		{
			name: "Test Positive GET - no withdrawals found",
			want: want{
				code:     204,
				response: nil,
			},
			args: arguments{
				url: "http://localhost:8080/api/user/withdrawals",
			},
		},
	}
	handler := &Handler{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	ur := &UserRouter{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	br := &BalanceRouter{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	handler.Post("/api/user/login", ur.Login)
	handler.Get("/api/user/withdrawals", br.GetWithdrawals)
	handler.Post("/api/user/register", ur.RegisterUser)
	ts := httptest.NewServer(handler)
	handler.Cursor.SaveUserInfo(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	for _, withdrawal := range mockWithdrawals {
		withdrawal.User = "test"
		handler.Cursor.SaveWithdrawal(withdrawal)
		withdrawal.User = ""
	}

	defer ts.Close()

	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	encoder.Encode(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/login", buff)
	request.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()

	cookies := res.Cookies()

	assert.Equal(t, res.StatusCode, 200)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.args.url, nil)

			w := httptest.NewRecorder()
			if tt.name == "Test Positive GET - no withdrawals found" {
				buff := bytes.NewBuffer([]byte{})
				encoder := json.NewEncoder(buff)
				encoder.Encode(&models.UserInfo{
					Username: "test2",
					Password: "test2",
				})
				request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/register", buff)
				request.Header.Add("Content-Type", "application/json")

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, request)
				res := w.Result()
				defer res.Body.Close()
				cookies = res.Cookies()
			}

			request.AddCookie(cookies[0])
			handler.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			receivedWithdrawals := []*models.Withdrawal{}
			if err := json.NewDecoder(res.Body).Decode(&receivedWithdrawals); err != nil {
				panic(err)
			}
			assert.Equal(t, tt.want.response, receivedWithdrawals)
		})
	}
}
