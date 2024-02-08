package benchmark

import (
	"fmt"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
)

func Benchmark(iter int, token, src, dest string) (time.Duration, error) {

	start := time.Now()
	client := tmdbapi.New(token, 5 * time.Second)
	for i := 0; i < iter; i++ {
		out, err := client.GetPath(src, dest)
		if err != nil { return 0, err }
		fmt.Println(out)
		/*err := runTest(token, src, dest)
		if err != nil { return 0, err }*/
	}
	dur := time.Since(start)

	return time.Duration(int(dur) / iter), nil
}

func runTest(token, src, dest string) error {
	client := tmdbapi.New(token, 5 * time.Second)
	
	out, err := client.GetPath(src, dest)
	fmt.Println(out)
	return err
}
