package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestFetchOne_Success(t *testing.T) {
	t.Parallel()
	userPayload := map[string]interface{}{
		"id":         1,
		"login":      "sample_username",
		"avatar_url": "https://",
		"html_url":   "https://",
		"url":        "https://",
		"type":       "User",
		"site_admin": false,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(userPayload)
	}))
	defer server.Close()

	client := &GitHubClient{
		httpClient:  server.Client(),
		accessToken: "",
		apiBaseURL:  server.URL,
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}

	user, err := client.FetchOne(context.Background(), "sample_username")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "sample_username", user.Login)
}

func TestFetchOne_NotFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &GitHubClient{
		httpClient:  server.Client(),
		accessToken: "",
		apiBaseURL:  server.URL,
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}

	user, err := client.FetchOne(context.Background(), "missing")
	require.Error(t, err)
	require.Nil(t, user)
}

func TestFetchUsersSince_UnexpectedStatus(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"error":"teapot"}`))
	}))
	defer server.Close()

	client := &GitHubClient{
		httpClient:  server.Client(),
		accessToken: "",
		apiBaseURL:  server.URL,
		rateLimiter: rate.NewLimiter(rate.Inf, 1),
	}

	users, err := client.FetchUsersSince(context.Background(), 0, 10)
	require.Error(t, err)
	require.Nil(t, users)
}
