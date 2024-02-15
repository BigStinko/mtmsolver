package benchmark

import (
	"fmt"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
)

type getPathFunc func(*tmdbapi.Client, string, string) ([]int, error)

func Benchmark(
	getPath getPathFunc,
	token, src, dest string,
	iter int, 
) error {
	fmt.Println("Running cache benchmark")
	start := time.Now()
	avg, err := runBenchCache(getPath, iter, token, src, dest)
	if err != nil { return err }
	dur := time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("Finished with average length %f, and average time of %s\n",
		avg, dur.String(),
	)

	fmt.Println("Running no cache benchmark")
	start = time.Now()
	avg, err = runBenchNoCache(getPath, iter, token, src, dest)
	if err != nil { return err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("Finished with average length %f, and average time of %s\n",
		avg, dur.String(),
	)
	
	return nil
}

func Compare(
	test1, test2 getPathFunc,
	test1Title, test2Title, token, src, dest string,
	iter int, 
) error {
	fmt.Printf("Running cache benchmark for \"%s\"...\n", test1Title)
	start := time.Now()
	avg, err := runBenchCache(test1, iter, token, src, dest)
	if err != nil { return err }
	dur := time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with average length %f, and average time of %s\n",
		test1Title, avg, dur.String(),
	)

	fmt.Printf("Running cache benchmark for \"%s\"...\n", test2Title)
	start = time.Now()
	avg, err = runBenchCache(test2, iter, token, src, dest)
	if err != nil { return err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with average length %f, and average time of %s\n",
		test2Title, avg, dur.String(),
	)

	fmt.Printf("Running no cache benchmark for \"%s\"...\n", test1Title)
	start = time.Now()
	avg, err = runBenchNoCache(test1, iter, token, src, dest)
	if err != nil { return err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with average length %f, and average time of %s\n",
		test1Title, avg, dur.String(),
	)

	fmt.Printf("Running no cache benchmark for \"%s\"...\n", test2Title)
	start = time.Now()
	avg, err = runBenchNoCache(test2, iter, token, src, dest)
	if err != nil { return err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with average length %f, and average time of %s\n",
		test2Title, avg, dur.String(),
	)
	
	return nil
}

func runBenchCache(
	getPath getPathFunc,
	iter int,
	token, src, dest string,
) (float64, error) {
	count := 0
	client := tmdbapi.New(token, time.Second * 5)
	for i := 0; i < iter; i++ {
		out, err := getPath(&client, src, dest)
		if err != nil { return 0, err }
		count += len(out) - 1
	}
	return float64(count) / float64(iter), nil
}

func runBenchNoCache(
	getPath getPathFunc,
	iter int,
	token, src, dest string,
) (float64, error) {
	count := 0
	for i := 0; i < iter; i++ {
		client := tmdbapi.New(token, time.Second * 5)
		out, err := getPath(&client, src, dest)
		if err != nil { return 0, err }
		count += len(out) - 1
	}
	return float64(count) / float64(iter), nil
}
