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
	err := benchmark.Benchmark(tmdbapi.GetPath, bearerToken, 1)
	if err != nil { fmt.Println(err) }
	/*avg1, avg2 := 0.0, 0.0
	for i := 0; i < 5; i++ {
		a1, a2, err := benchmark.Compare(
			tmdbapi.GetPath3, tmdbapi.GetPath2,
			"GetPath3", "GetPath2", bearerToken, 10)
		avg1 += a1
		avg2 += a2
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Printf("success rate for getpath is %f\nsuccess rate for getpathexperimental is %f\n", avg1 / 5.0, avg2 / 5.0)*/
}
