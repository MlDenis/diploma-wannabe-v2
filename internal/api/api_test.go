package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/MlDenis/diploma-wannabe-v2/internal/db"
	"github.com/MlDenis/diploma-wannabe-v2/internal/mocks"
	"github.com/MlDenis/diploma-wannabe-v2/internal/models"
)

func TestCookies(t *testing.T) {
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
	handler.Post("/api/user/register", ur.RegisterUser)
	handler.Post("/api/user/login", ur.Login)
	ts := httptest.NewServer(handler)

	defer ts.Close()

	buff := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buff)
	err := encoder.Encode(&models.UserInfo{
		Username: "test",
		Password: "test",
	})
	if err != nil {
		return
	}
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/user/register", buff)
	request.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, request)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	assert.Equal(t, res.StatusCode, 200)
}
