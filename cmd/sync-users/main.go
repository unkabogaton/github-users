package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/unkabogaton/github-users/internal/application/cache"
	"github.com/unkabogaton/github-users/internal/domain/entities"
	"github.com/unkabogaton/github-users/internal/infrastructure/database/repositories"
	"github.com/unkabogaton/github-users/internal/infrastructure/http"
)

func convertEnvConfigToInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func main() {
	_ = godotenv.Load()

	applicationContext := context.Background()

	var (
		usersPerPage            = convertEnvConfigToInt("USERS_PER_PAGE", 30)
		workerPoolSize          = convertEnvConfigToInt("WORKER_POOL_SIZE", 5)
		maximumFetchRetries     = convertEnvConfigToInt("MAXIMUM_FETCH_RETRIES", 3)
		delayBetweenUpsertsMS   = convertEnvConfigToInt("DELAY_BETWEEN_UPSERTS_MS", 200)
		maximumConsecutiveEmpty = convertEnvConfigToInt("MAXIMUM_CONSECUTIVE_EMPTY", 1)
	)

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	database, databaseErr := sqlx.Open("mysql", dsn)
	if databaseErr != nil {
		panic(fmt.Errorf("failed to open database: %w", databaseErr))
	}
	defer database.Close()

	userRepository := repositories.NewUserRepository(database)

	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisTTLString := os.Getenv("REDIS_TTL_SEC")
	if redisTTLString == "" {
		redisTTLString = "300"
	}
	redisTTLSeconds, _ := strconv.Atoi(redisTTLString)
	_ = cache.NewRedisCache(redisAddress, redisPassword, redisTTLSeconds)

	gitHubAccessToken := os.Getenv("GITHUB_TOKEN")
	gitHubClient := http.NewGitHubClient(gitHubAccessToken)

	userChannel := make(chan entities.User, usersPerPage*workerPoolSize)
	var workerWaitGroup sync.WaitGroup

	for workerIndex := 0; workerIndex < workerPoolSize; workerIndex++ {
		workerWaitGroup.Add(1)

		go func() {
			defer workerWaitGroup.Done()
			for userRecord := range userChannel {
				if err := userRepository.Upsert(applicationContext, &userRecord); err != nil {
					fmt.Fprintf(
						os.Stderr,
						"upsert error (login %s, id %d): %v\n",
						userRecord.Login,
						userRecord.ID,
						err,
					)
				}
				time.Sleep(time.Duration(delayBetweenUpsertsMS) * time.Millisecond)
			}
		}()
	}

	go func() {
		defer close(userChannel)

		lastFetchedID := 0
		consecutiveEmptyBatches := 0

		for {
			var lastFetchError error
			var fetchedUsers []entities.GitHubUser

			for attempt := 1; attempt <= maximumFetchRetries; attempt++ {
				usersBatch, fetchErr := gitHubClient.FetchUsersSince(
					applicationContext,
					lastFetchedID,
					usersPerPage,
				)
				if fetchErr != nil {
					lastFetchError = fetchErr
					time.Sleep(time.Duration(attempt) * time.Second)
					continue
				}
				fetchedUsers = usersBatch
				lastFetchError = nil
				break
			}

			if lastFetchError != nil {
				fmt.Fprintf(
					os.Stderr,
					"fetch failed after %d attempts (since=%d): %v\n",
					maximumFetchRetries,
					lastFetchedID,
					lastFetchError,
				)
				return
			}

			if len(fetchedUsers) == 0 {
				consecutiveEmptyBatches++
				if consecutiveEmptyBatches >= maximumConsecutiveEmpty {
					return
				}
				continue
			}
			consecutiveEmptyBatches = 0

			nextSinceID := lastFetchedID
			for _, fetchedUser := range fetchedUsers {
				if fetchedUser.ID > nextSinceID {
					nextSinceID = fetchedUser.ID
				}
				userChannel <- entities.User{
					ID:           fetchedUser.ID,
					Login:        fetchedUser.Login,
					NodeID:       fetchedUser.NodeID,
					AvatarURL:    fetchedUser.AvatarURL,
					HTMLURL:      fetchedUser.HTMLURL,
					URL:          fetchedUser.URL,
					Type:         fetchedUser.Type,
					UserViewType: fetchedUser.UserViewType,
					SiteAdmin:    fetchedUser.SiteAdmin,
				}
			}
			lastFetchedID = nextSinceID
		}
	}()

	workerWaitGroup.Wait()
	fmt.Println("GitHub user synchronization complete.")
}
