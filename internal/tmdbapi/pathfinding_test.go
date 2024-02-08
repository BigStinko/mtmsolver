package tmdbapi

import (
	"net/http"
	"testing"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbcache"
)

func TestPathfinding(t *testing.T) {
	var e = struct{}{}
	cache := tmdbcache.Test(
		map[int]map[int]struct{}{
			1037: {500:e, 14839:e, 680:e},
			3129: {500:e, 680:e, 1724:e},
			147: {500:e, 393:e, 24:e},
			2969: {500:e, 2109:e},
		},
		map[int]map[int]struct{}{
			500: {1037:e, 3129:e, 147:e, 2969:e},
			14839: {1037:e},
			680: {1037:e, 3129:e},
			1724: {3129:e},
			393: {147:e},
			24: {147:e},
			2109: {2969:e},
		},
		make(map[int]map[int]struct{}),
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
