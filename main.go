package main

import (
	"fmt"
	"os"
	_"time"

	"github.com/BigStinko/mtmsolver/internal/benchmark"
	_"github.com/BigStinko/mtmsolver/internal/tmdbapi"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bearerToken := os.Getenv("BEARER_TOKEN")
	//client := tmdbapi.New("Bearer " + bearerToken, time.Second * 5)
	//out, err := client.GetPath("The City of Lost Children", "Empire of the Sun")
	//if err != nil { fmt.Println(err.Error()) }
	//client.PrintPath(out)

	err := benchmark.Benchmark(150, "Bearer " + bearerToken, "Midsommar", "Gravity")
	//dur, err := benchmark.Benchmark(900, "Bearer " + bearerToken, "Reservoir Dogs", "Pulp Fiction")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	
}
