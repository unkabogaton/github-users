package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/unkabogaton/github-users/internal/application/cache"
	"github.com/unkabogaton/github-users/internal/application/services"
	"github.com/unkabogaton/github-users/internal/infrastructure/database/repositories"
	"github.com/unkabogaton/github-users/internal/infrastructure/http"
	"github.com/unkabogaton/github-users/internal/infrastructure/http/controllers"
	"github.com/unkabogaton/github-users/internal/infrastructure/http/middleware"
)

func main() {
	_ = godotenv.Load()
	restServerAddress := os.Getenv("REST_ADDRESS")
	if restServerAddress == "" {
		restServerAddress = ":8080"
	}

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
	redisCache := cache.NewRedisCache(redisAddress, redisPassword, redisTTLSeconds)

	gitHubToken := os.Getenv("GITHUB_TOKEN")
	gitHubClient := http.NewGitHubClient(gitHubToken)
	userService := services.NewUserService(userRepository, redisCache, gitHubClient)

	router := gin.Default()
	router.SetTrustedProxies(nil)
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
