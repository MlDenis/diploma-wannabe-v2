package api

import (
	"bytes"
	"encoding/json"
	"github.com/MlDenis/diploma-wannabe-v2/internal/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
)

func TestBalanceGet(t *testing.T) {
	expectedBalance := &models.Balance{
		Current:   500.5,
		Withdrawn: 42,
	}
	type want struct {
		code int
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
			name: "Test Positive balance get",
			want: want{
				code: 200,
			},
			args: arguments{
				url: "http://localhost:8080/api/user/balance",
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
	handler.Post("/api/user/register", ur.RegisterUser)
	handler.Post("/api/user/login", ur.Login)
	handler.Get("/api/user/balance", br.GetBalance)
	ts := httptest.NewServer(handler)
	err := handler.Cursor.SaveUserInfo(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		return
	}

	result, _ := handler.Cursor.UpdateUserBalance(
		"test", expectedBalance,
	)
	assert.Equal(t, expectedBalance, result)

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

	assert.Equal(t, res.StatusCode, 200)

	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.args.url, nil)
			request.AddCookie(cookies[0])
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(res.Body)

			actualBalance := &models.Balance{}
			if err := json.NewDecoder(res.Body).Decode(&actualBalance); err != nil {
				panic(err)
			}
			assert.Equal(t, expectedBalance, actualBalance)

		})
	}
}
