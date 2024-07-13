package test_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
	"user-service/internal/model"
)

const (
	test_http_timeout      = time.Second
	user_service_address   = "http://127.0.0.1:8080"
	user_service_url_users = user_service_address + "/v1/users"
	user_service_url_user  = user_service_url_users + "/%s"
)

type ErrResponse struct {
	Error string `json:"error"`
}

func CallCreateUserEndpoint(t *testing.T, u model.User) ([]byte, int) {
	userBytes, err := json.Marshal(u)
	require.NoError(t, err)

	return callEndpoint(t, userBytes, http.MethodPost, user_service_url_users)
}

func CallUpdateUserEndpoint(t *testing.T, u model.User) ([]byte, int) {
	userBytes, err := json.Marshal(u)
	require.NoError(t, err)

	userURL := fmt.Sprintf(user_service_url_user, u.ID.String())
	return callEndpoint(t, userBytes, http.MethodPut, userURL)
}

func CallDeleteUserEndpoint(t *testing.T, userID uuid.UUID) ([]byte, int) {
	userURL := fmt.Sprintf(user_service_url_user, userID.String())
	return callEndpoint(t, nil, http.MethodDelete, userURL)
}

func CallGetUserEndpoint(t *testing.T, userID uuid.UUID) ([]byte, int) {
	userURL := fmt.Sprintf(user_service_url_user, userID.String())
	return callEndpoint(t, nil, http.MethodGet, userURL)
}

func CallPath(t *testing.T, method, path string) ([]byte, int) {
	url := user_service_address + path
	return callEndpoint(t, nil, method, url)
}

func callEndpoint(t *testing.T, payload []byte, method, url string) ([]byte, int) {
	ctx, cancel := context.WithTimeout(context.Background(), test_http_timeout)
	defer cancel()

	var reader io.Reader
	if len(payload) != 0 {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	require.NoError(t, err, "request creation failed")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "http request failed")
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	return respBytes, resp.StatusCode
}
