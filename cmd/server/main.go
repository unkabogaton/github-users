package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/unkabogaton/github-users/internal/cache"
	"github.com/unkabogaton/github-users/internal/clients"
	"github.com/unkabogaton/github-users/internal/repositories"
	"github.com/unkabogaton/github-users/internal/rest"
	"github.com/unkabogaton/github-users/internal/service"
)

func main() {
	restServerAddress := os.Getenv("REST_ADDRESS")
	if restServerAddress == "" {
		restServerAddress = ":8080"
	}

	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		log.Fatal("POSTGRES_DSN environment variable is required")
	}
	database, databaseErr := sqlx.Open("postgres", postgresDSN)
	if databaseErr != nil {
		panic(fmt.Errorf("failed to open database: %w", databaseErr))
	}
	defer database.Close()
	userRepository := repositories.NewSQLXUserRepo(database)

	// Prefer REDIS_ADDRESS, fallback to REDIS_ADDR
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = os.Getenv("REDIS_ADDR")
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisTTLString := os.Getenv("REDIS_TTL_SEC")
	if redisTTLString == "" {
		redisTTLString = "300"
	}
	redisTTLSeconds, _ := strconv.Atoi(redisTTLString)
	redisCache := cache.NewRedisCache(redisAddress, redisPassword, redisTTLSeconds)

	gitHubToken := os.Getenv("GITHUB_TOKEN")
	gitHubClient := clients.NewGitHubClient(gitHubToken, userRepository, redisCache)
	userService := service.NewUserService(userRepository, redisCache, gitHubClient)

	router := gin.Default()
	userHandler := rest.NewHandler(userService)

	router.GET("/users", userHandler.ListUsers)
	router.PUT("/users/:username", userHandler.UpdateUser)
	router.GET("/users/:username", userHandler.GetUser)
	router.DELETE("/users/:username", userHandler.DeleteUser)

	log.Printf("Starting REST server on %s", restServerAddress)
	if runError := router.Run(restServerAddress); runError != nil {
		log.Fatalf("server failed: %v", runError)
	}
}
