package main

import (
	"fmt"
	"os"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bearerToken := os.Getenv("BEARER_TOKEN")
	client := tmdbapi.New("Bearer " + bearerToken, 5 * time.Second)

	err := client.GetPath("The Iron Claw", "High School Musical")
	if err != nil {
		fmt.Print(err.Error())
	}
}
