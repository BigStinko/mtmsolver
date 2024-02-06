package tmdbapi

import (
	"net/http"
	"testing"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbcache"
)

func TestPathfinding(t *testing.T) {
	cache := tmdbcache.Test(
		map[int][]int{
			1037: {500, 14839, 680},
			3129: {500, 680, 1724},
			147: {500, 393, 24},
			2969: {500, 2109},
		},
		map[int][]int{
			500: {1037, 3129, 147, 2969},
			14839: {1037},
			680: {1037, 3129},
			1724: {3129},
			393: {147},
			24: {147},
			2109: {2969},
		},
		map[int][]int{},
	)

	client := Client{
		httpClient: http.Client{
			Timeout: time.Second * 5,
		},
		authHeader: "",
		cache: cache,
	}

	path, err := client.bfs(500, 680)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(path) != 2 {
		t.Fatalf("Incorrect response got %v", path)
	}

	if path[0] != 500 && path[1] != 680 {
		t.Fatalf("Incorrect response got %v", path)
	}
}
