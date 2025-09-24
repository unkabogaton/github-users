package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/unkabogaton/github-users/internal/application/cache"
	"github.com/unkabogaton/github-users/internal/application/services"
	"github.com/unkabogaton/github-users/internal/infrastructure/database/repositories"
	"github.com/unkabogaton/github-users/internal/infrastructure/http"
	"github.com/unkabogaton/github-users/internal/infrastructure/http/controllers"
    "github.com/unkabogaton/github-users/internal/infrastructure/http/middleware"
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
	userRepository := repositories.NewUserRepository(database)

	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisTTLString := os.Getenv("REDIS_TTL_SEC")
	if redisTTLString == "" {
		redisTTLString = "300"
	}
	redisTTLSeconds, _ := strconv.Atoi(redisTTLString)
	redisCache := cache.NewRedisCache(redisAddress, redisPassword, redisTTLSeconds)

	gitHubToken := os.Getenv("GITHUB_TOKEN")
	gitHubClient := http.NewGitHubClient(gitHubToken)
	userService := services.NewUserService(userRepository, redisCache, gitHubClient)

    router := gin.Default()
    router.Use(middleware.ErrorHandlingMiddleware())
	userController := controllers.NewUserController(userService)

	router.GET("/users", userController.ListUsers)
	router.PUT("/users/:username", userController.UpdateUser)
	router.GET("/users/:username", userController.GetUser)
	router.DELETE("/users/:username", userController.DeleteUser)

	log.Printf("Starting REST server on %s", restServerAddress)
	if runError := router.Run(restServerAddress); runError != nil {
		log.Fatalf("server failed: %v", runError)
	}
}
