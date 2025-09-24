package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/unkabogaton/github-users/internal/application/cache"
	"github.com/unkabogaton/github-users/internal/application/services"
	"github.com/unkabogaton/github-users/internal/infrastructure/database/repositories"
	grpcserver "github.com/unkabogaton/github-users/internal/infrastructure/grpc"
	httpclient "github.com/unkabogaton/github-users/internal/infrastructure/http"
)

func main() {
	_ = godotenv.Load()

	grpcAddress := os.Getenv("GRPC_ADDRESS")
	if grpcAddress == "" {
		grpcAddress = ":9090"
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
		log.Fatalf("failed to open database: %v", databaseErr)
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
	gitHubClient := httpclient.NewGitHubClient(gitHubToken)
	userService := services.NewUserService(userRepository, redisCache, gitHubClient)

	server := grpcserver.NewServer(userService)
	log.Printf("Starting gRPC server on %s", grpcAddress)
	if err := server.ListenAndServe(grpcAddress); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
