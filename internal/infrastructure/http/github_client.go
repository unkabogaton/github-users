package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/domain/interfaces"

	"golang.org/x/time/rate"
)

type GitHubClient struct {
	httpClient  *http.Client
	accessToken string
	apiBaseURL  string
	rateLimiter *rate.Limiter
}

func NewGitHubClient(accessToken string) interfaces.GitHubClient {
	rateLimiter := rate.NewLimiter(rate.Limit(1), 2)

	return &GitHubClient{
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		accessToken: accessToken,
		apiBaseURL:  "https://api.github.com",
		rateLimiter: rateLimiter,
	}
}

func (c *GitHubClient) FetchUsersSince(
	ctx context.Context,
	lastUserID int,
	resultsPerPage int,
) ([]entities.GitHubUser, error) {

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	requestURL := fmt.Sprintf(
		"%s/users?per_page=%d&since=%d",
		c.apiBaseURL,
		resultsPerPage,
		lastUserID,
	)

	httpRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if c.accessToken != "" {
		httpRequest.Header.Set("Authorization", "token "+c.accessToken)
	}
	httpRequest.Header.Set("Accept", "application/vnd.github.v3+json")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == http.StatusTooManyRequests || httpResponse.StatusCode >= http.StatusInternalServerError {
		responseBody, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("GitHub API rate/server error: %d body: %s", httpResponse.StatusCode, string(responseBody))
	}
	if httpResponse.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(httpResponse.Body)
		return nil, fmt.Errorf("unexpected GitHub status %d: %s", httpResponse.StatusCode, string(responseBody))
	}

	var fetchedUsers []entities.GitHubUser
	if err := json.NewDecoder(httpResponse.Body).Decode(&fetchedUsers); err != nil {
		return nil, err
	}

	return fetchedUsers, nil
}

func (c *GitHubClient) FetchOne(
	ctx context.Context,
	username string,
) (*entities.GitHubUser, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	userRequestURL := fmt.Sprintf("%s/users/%s", c.apiBaseURL, username)

	httpRequest, _ := http.NewRequestWithContext(ctx, http.MethodGet, userRequestURL, nil)
	if c.accessToken != "" {
		httpRequest.Header.Set("Authorization", "token "+c.accessToken)
	}
	httpRequest.Header.Set("Accept", "application/vnd.github.v3+json")

	httpResponse, err := c.httpClient.Do(httpRequest)
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

	var fetchedUser entities.GitHubUser
	if err := json.NewDecoder(httpResponse.Body).Decode(&fetchedUser); err != nil {
		return nil, err
	}

	return &fetchedUser, nil
}
