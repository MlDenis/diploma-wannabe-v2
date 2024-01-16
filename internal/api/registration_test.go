package api

import (
	"bytes"
	"encoding/json"
	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegistration(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type userinfo struct {
		Username string `json:"login"`
		Password string `json:"password"`
	}
	type arguments struct {
		url     string
		payload userinfo
	}
	tests := []struct {
		name string
		want want
		args arguments
	}{
		{
			name: "Test Positive registration",
			want: want{
				code:     200,
				response: `user created successfully`,
			},
			args: arguments{
				url: "http://localhost:8080/api/user/register",
				payload: userinfo{
					Username: "test",
					Password: "test",
				},
			},
		},
		{
			name: "Test Negative registration user exists",
			want: want{
				code:     409,
				response: "user already exists\n",
			},
			args: arguments{
				url: "http://localhost:8080/api/user/register",
				payload: userinfo{
					Username: "test",
					Password: "test",
				},
			},
		},
		{
			name: "Test Negative registration invalid request",
			want: want{
				code:     400,
				response: "validation error\n",
			},
			args: arguments{
				url:     "http://localhost:8080/api/user/register",
				payload: userinfo{},
			},
		},
		{
			name: "Test Negative registration user exists and different password",
			want: want{
				code:     409,
				response: "user already exists\n",
			},
			args: arguments{
				url: "http://localhost:8080/api/user/register",
				payload: userinfo{
					Username: "test",
					Password: "new test",
				},
			},
		},
	}
	ur := &UserRouter{
		Mux: chi.NewMux(),
		Cursor: &db.Cursor{
			IDBInterface: mocks.NewMock(),
		},
	}
	ur.Post("/api/user/register", ur.RegisterUser)
	ts := httptest.NewServer(ur)

	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff := bytes.NewBuffer([]byte{})
			encoder := json.NewEncoder(buff)
			encoder.Encode(&tt.args.payload)
			request := httptest.NewRequest(http.MethodPost, tt.args.url, buff)
			request.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			ur.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}
		})
	}
}
