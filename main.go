package main

import (
	"os"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bearerToken := os.Getenv("BEARER_TOKEN")
	client := tmdbapi.NewClient("Bearer " + bearerToken, 5 * time.Second, 5 * time.Minute)
}
