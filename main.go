package main

import (
	"fmt"
	"os"
	_ "time"

	"github.com/BigStinko/mtmsolver/internal/benchmark"
	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
	_ "github.com/BigStinko/mtmsolver/internal/tmdbapi"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bearerToken := "Bearer " + os.Getenv("BEARER_TOKEN")
	//client := tmdbapi.New("Bearer " + bearerToken, time.Second * 5)
	//out, err := client.GetPath("The City of Lost Children", "Empire of the Sun")
	//if err != nil { fmt.Println(err.Error()) }
	//client.PrintPath(out)
	//err := benchmark.Benchmark(tmdbapi.GetPathExperimental, bearerToken, "Midsommar", "Gravity", 100)
	err := benchmark.Compare(
		tmdbapi.GetPath, tmdbapi.GetPathExperimental,
		"GetPath", "GetPathExperimental",
		bearerToken, "Midsommar", "Gravity", 75,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
