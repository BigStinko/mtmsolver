package benchmark

import (
	"fmt"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
)

func Benchmark(iter int, token, src, dest string) error {
	start := time.Now()

	fmt.Println("running benchmark async with cache")
	start = time.Now()
	count := 0
	client := tmdbapi.New(token, 5 * time.Second)
	for i := 0; i < iter; i++ {
		out, err := client.GetPath(src, dest)
		if err != nil { return err }
		count += len(out) - 1
		fmt.Println(out)
	}
	dur := time.Since(start)
	fmt.Println(dur.String())
	fmt.Printf("average length of path: %f\n", float32(count)/float32(iter))

	fmt.Println("running benchmark async")
	start = time.Now()
	count = 0
	for i := 0; i < iter; i++ {
		client := tmdbapi.New(token, 5 * time.Second)
		out, err := client.GetPath(src, dest)
		if err != nil { return err }
		count += len(out) - 1
		fmt.Println(out)
	}
	dur = time.Since(start)
	fmt.Println(dur.String())
	fmt.Printf("average length of path: %f\n", float32(count)/float32(iter))

	return nil
}

func runTest(token, src, dest string) error {
	client := tmdbapi.New(token, 5 * time.Second)
	
	out, err := client.GetPath(src, dest)
	fmt.Println(out)
	return err
}
