package benchmark

import (
	"fmt"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbapi"
)

type getPathFunc func(*tmdbapi.Client, string, string) ([]int, error)



func Benchmark(getPath getPathFunc, token string, iter int) error {
	fmt.Println("Running cache benchmark")
	start := time.Now()
	total := 0.0
	for i := 0; i < iter; i++ {
		avg, err := runBenchCache(getPath, token)
		if err != nil { return err }
		total += avg
	}
	dur := time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\nFinished with success rate %f, and average time of %s\n",
		total / float64(iter), dur.String(),
	)

	fmt.Println("Running no cache benchmark")
	start = time.Now()
	total = 0.0
	for i := 0; i < iter; i++ {
		avg, err := runBenchNoCache(getPath, token)
		if err != nil { return err }
		total += avg
	}
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\nFinished with success rate %f, and average time of %s\n",
		total / float64(iter), dur.String(),
	)
	
	return nil
}

func Compare(
	test1, test2 getPathFunc,
	test1Title, test2Title, token string,
	iter int, 
) (float64, float64, error) {
	avgLen1, avgLen2 := 0.0, 0.0
	fmt.Printf("Running cache benchmark for \"%s\"...\n", test1Title)
	start := time.Now()
	avg, err := runBenchCache(test1, token)
	avgLen1 += avg
	if err != nil { return 0.0, 0.0, err }
	dur := time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with success rate %f, and average time of %s\n",
		test1Title, avg, dur.String(),
	)

	fmt.Printf("Running cache benchmark for \"%s\"...\n", test2Title)
	start = time.Now()
	avg, err = runBenchCache(test2, token)
	avgLen2 += avg
	if err != nil { return 0.0, 0.0, err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with success rate %f, and average time of %s\n",
		test2Title, avg, dur.String(),
	)

	fmt.Printf("Running no cache benchmark for \"%s\"...\n", test1Title)
	start = time.Now()
	avg, err = runBenchNoCache(test1, token)
	avgLen1 += avg
	if err != nil { return 0.0, 0.0, err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with success rate %f, and average time of %s\n",
		test1Title, avg, dur.String(),
	)

	fmt.Printf("Running no cache benchmark for \"%s\"...\n", test2Title)
	start = time.Now()
	avg, err = runBenchNoCache(test2, token)
	avgLen2 += avg
	if err != nil { return 0.0, 0.0, err }
	dur = time.Since(start)
	dur /= time.Duration(iter)
	fmt.Printf("\"%s\" Finished with success rate %f, and average time of %s\n",
		test2Title, avg, dur.String(),
	)
	
	return avgLen1 / 2.0, avgLen2 / 2.0 , nil
}

func runBenchCache(
	getPath getPathFunc,
	token string,
) (float64, error) {
	tests := map[int]struct{
		src string
		dest string
		expectedLength int
	}{
		0: {
			src: "Reservoir Dogs",
			dest: "Pulp Fiction",
			expectedLength: 1,
		},
		1: {
			src: "The City of Lost Children",
			dest: "Empire of the Sun",
			expectedLength: 2,
		},
		2: {
			src: "Midsommar",
			dest: "Gravity",
			expectedLength: 3,
		},
		3: {
			src: "The Descent",
			dest: "Prisoners",
			expectedLength: 2,
		},
		4: {
			src: "Fight Club",
			dest: "Rounders",
			expectedLength: 1,
		},
		5: {
			src: "Kickboxer",
			dest: "Dirty Rotten Scoundrels",
			expectedLength: 3,
		},
	}
	count := 0
	client := tmdbapi.New(token, time.Second * 5)
	for _, test := range tests {
		out, err := getPath(&client, test.src, test.dest)
		fmt.Print(".")
		if err != nil { return 0, err }
		if len(out) - 1 == test.expectedLength {
			count++
		}
	}
	return float64(count) / 6.0, nil
}

func runBenchNoCache(getPath getPathFunc, token string) (float64, error) {
	tests := map[int]struct{
		src string
		dest string
		expectedLength int
	}{
		0: {
			src: "Reservoir Dogs",
			dest: "Pulp Fiction",
			expectedLength: 1,
		},
		1: {
			src: "The City of Lost Children",
			dest: "Empire of the Sun",
			expectedLength: 2,
		},
		2: {
			src: "Midsommar",
			dest: "Gravity",
			expectedLength: 3,
		},
		3: {
			src: "The Descent",
			dest: "Prisoners",
			expectedLength: 2,
		},
		4: {
			src: "Fight Club",
			dest: "Rounders",
			expectedLength: 1,
		},
		5: {
			src: "Kickboxer",
			dest: "Dirty Rotten Scoundrels",
			expectedLength: 3,
		},
	}
	count := 0
	for _, test := range tests {
		client := tmdbapi.New(token, time.Second * 5)
		out, err := getPath(&client, test.src, test.dest)
		fmt.Print(".")
		if err != nil { return 0, err }
		if len(out) - 1 == test.expectedLength {
			count++
		}
	}
	return float64(count) / 6.0, nil
}
