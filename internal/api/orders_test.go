package api

import (
	"bytes"
	"encoding/json"
	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/logger"
	"github.com/MlDenis/diploma-wannabe-v2/internal/mocks"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPostOrders(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type arguments struct {
		url         string
		number      interface{}
		contentType string
	}
	tests := []struct {
		name string
		want want
		args arguments
	}{
		{
			name: "Test Positive order post",
			want: want{
				code:     202,
				response: "new order created",
			},
			args: arguments{
				url:         "http://localhost:8080/api/user/orders",
				number:      "14242412",
				contentType: "text/plain",
			},
		},
		{
			name: "Test Positive order posted already",
			want: want{
				code:     200,
				response: "order created already",
			},
			args: arguments{
				url:         "http://localhost:8080/api/user/orders",
				number:      "14242412",
				contentType: "text/plain",
			},
		},
		{
			name: "Test Negative post order wrong number",
			want: want{
				code:     422,
				response: "wrong number format\n",
			},
			args: arguments{
				url:         "http://localhost:8080/api/user/orders",
				number:      "714683",
				contentType: "text/plain",
			},
		},
		{
			name: "Test Negative post order already registered by another user",
			want: want{
				code:     409,
				response: "order was uploaded already by another user\n",
			},
			args: arguments{
				url:         "http://localhost:8080/api/user/orders",
				number:      "14242412",
				contentType: "text/plain",
			},
		},
		{
			name: "Test Negative post order bad request",
			want: want{
				code:     400,
				response: "wrong content\n",
			},
			args: arguments{
				url:         "http://localhost:8080/api/user/orders",
				number:      "0",
				contentType: "application/json",
			},
		},
	}

	l, _ := logger.InitializeLogger("info")

	handler := &Handler{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	r := &OrderRouter{
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

	handler.Post("/api/user/register", ur.RegisterUser)
	handler.Post("/api/user/login", ur.Login)
	handler.Post("/api/user/orders", r.UploadOrder)
	ts := httptest.NewServer(handler)

	handler.Cursor.SaveUserInfo(&models.UserInfo{
		Username: "test",
		Password: "test",
	}, l)

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
			_, _ = buff.Write([]byte(tt.args.number.(string)))
			request := httptest.NewRequest(http.MethodPost, tt.args.url, buff)
			request.Header.Add("Content-Type", tt.args.contentType)

			w := httptest.NewRecorder()
			if tt.name == "Test Negative post order already registered by another user" {
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

func TestGetOrders(t *testing.T) {
	type want struct {
		code     int
		response string
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
			name: "Test Positive get order no data found",
			want: want{
				code:     204,
				response: "no orders foun",
			},
			args: arguments{
				url: "http://localhost:8080/api/user/orders",
			},
		},
		{
			name: "Test Positive order get",
			want: want{
				code:     200,
				response: `[{"number":"9278923470","status":"PROCESSED","accrual":500,"uploaded_at":"2020-12-10T15:15:45+03:00"},{"number":"12345678903","status":"PROCESSING","uploaded_at":"2020-12-10T15:12:01+03:00"},{"number":"346436439","status":"INVALID","uploaded_at":"2020-12-09T16:09:53+03:00"},{"number":"3464364393333","status":"NEW","uploaded_at":"2020-12-09T16:09:53+03:00"}]`,
			},
			args: arguments{
				url: "http://localhost:8080/api/user/orders",
			},
		},
	}
	layout := "2006-01-02T15:04:05Z07:00"
	parseTime := func(layout string, toParse string) time.Time {
		parsed, _ := time.Parse(layout, toParse)
		return parsed
	}
	orders := []*models.Order{
		{
			Number:     "9278923470",
			Status:     "PROCESSED",
			Accrual:    500,
			UploadedAt: parseTime(layout, "2020-12-10T15:15:45+03:00"),
		},
		{
			Number:     "12345678903",
			Status:     "PROCESSING",
			UploadedAt: parseTime(layout, "2020-12-10T15:12:01+03:00"),
		},
		{
			Number:     "346436439",
			Status:     "INVALID",
			UploadedAt: parseTime(layout, "2020-12-09T16:09:53+03:00"),
		},
		{
			Number:     "3464364393333",
			Status:     "NEW",
			UploadedAt: parseTime(layout, "2020-12-09T16:09:53+03:00"),
		},
	}
	handler := &Handler{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	r := &OrderRouter{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	handler.Get("/api/user/orders", r.GetOrders)
	ts := httptest.NewServer(handler)

	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := logger.InitializeLogger("info")
			request := httptest.NewRequest(http.MethodGet, tt.args.url, nil)

			w := httptest.NewRecorder()
			if tt.name == "Test Positive order get" {
				for _, order := range orders {
					handler.Cursor.SaveOrder(order, l)
				}
			}
			handler.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.want.response, string(resBody)[:len(string(resBody))-1])
		})
	}
}
