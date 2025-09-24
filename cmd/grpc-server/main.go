//go:build grpc

package main

import (
    "log"
    "os"
    "strconv"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/joho/godotenv"

    "github.com/unkabogaton/github-users/internal/application/cache"
    "github.com/unkabogaton/github-users/internal/application/services"
    "github.com/unkabogaton/github-users/internal/infrastructure/database/repositories"
    httpclient "github.com/unkabogaton/github-users/internal/infrastructure/http"
    grpcserver "github.com/unkabogaton/github-users/internal/infrastructure/grpc"
)

func main() {
    _ = godotenv.Load()

    grpcAddress := os.Getenv("GRPC_ADDRESS")
    if grpcAddress == "" {
        grpcAddress = ":9090"
    }

    postgresDSN := os.Getenv("POSTGRES_DSN")
    if postgresDSN == "" {
        log.Fatal("POSTGRES_DSN environment variable is required")
    }
    database, databaseErr := sqlx.Open("postgres", postgresDSN)
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


