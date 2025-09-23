package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/unkabogaton/github-users/internal/cache"
	"github.com/unkabogaton/github-users/internal/models"
	"github.com/unkabogaton/github-users/internal/repositories"

	"golang.org/x/time/rate"
)

type GitHubClient struct {
	httpClient     *http.Client
	accessToken    string
	apiBaseURL     string
	userRepository repositories.UserRepo
	redisCache     *cache.RedisCache
	rateLimiter    *rate.Limiter
}

func NewGitHubClient(
	accessToken string,
	userRepository repositories.UserRepo,
	redisCache *cache.RedisCache,
) *GitHubClient {
	rateLimiter := rate.NewLimiter(rate.Limit(1), 2)

	return &GitHubClient{
		httpClient:     &http.Client{Timeout: 15 * time.Second},
		accessToken:    accessToken,
		apiBaseURL:     "https://api.github.com",
		userRepository: userRepository,
		redisCache:     redisCache,
		rateLimiter:    rateLimiter,
	}
}

func (client *GitHubClient) FetchUsersSince(
	ctx context.Context,
	lastUserID int,
	resultsPerPage int,
) ([]models.GitHubUser, error) {

	if err := client.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	requestURL := fmt.Sprintf(
		"%s/users?per_page=%d&since=%d",
		client.apiBaseURL,
		resultsPerPage,
		lastUserID,
	)

	httpRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if client.accessToken != "" {
		httpRequest.Header.Set("Authorization", "token "+client.accessToken)
	}
	httpRequest.Header.Set("Accept", "application/vnd.github.v3+json")

	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == http.StatusTooManyRequests || httpResponse.StatusCode >= 500 {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("GitHub API rate/server error: %d body: %s", httpResponse.StatusCode, string(body))
	}
	if httpResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unexpected GitHub status %d: %s", httpResponse.StatusCode, string(body))
	}

	var fetchedUsers []models.GitHubUser
	if err := json.NewDecoder(httpResponse.Body).Decode(&fetchedUsers); err != nil {
		return nil, err
	}

	return fetchedUsers, nil
}

func (client *GitHubClient) FetchOne(
	ctx context.Context,
	username string,
) (*models.GitHubUser, error) {
	if err := client.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	userRequestURL := fmt.Sprintf("%s/users/%s", client.apiBaseURL, username)

	httpRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, userRequestURL, nil)
	if client.accessToken != "" {
		httpRequest.Header.Set("Authorization", "token "+client.accessToken)
	}
	httpRequest.Header.Set("Accept", "application/vnd.github.v3+json")

	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == http.StatusTooManyRequests || httpResponse.StatusCode >= http.StatusInternalServerError {
		responseBody, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("GitHub API rate/server error: %d body: %s", httpResponse.StatusCode, string(responseBody))
	}

	if httpResponse.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user %s not found", username)
	}

	if httpResponse.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unexpected GitHub status %d: %s", httpResponse.StatusCode, string(responseBody))
	}

	var fetchedUser models.GitHubUser
	if err := json.NewDecoder(httpResponse.Body).Decode(&fetchedUser); err != nil {
		return nil, err
	}

	return &fetchedUser, nil
}
