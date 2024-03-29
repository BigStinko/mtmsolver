package tmdbapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BigStinko/mtmsolver/internal/tmdbcache"
)

type Client struct {
	httpClient http.Client
	cache      tmdbcache.Cache
	authHeader string
	searchFactor int
	maxRoutines int
}

type resource interface {
	ActorResource | ActorQueryResult |
	MovieResource | MovieQueryResult |
	Credits 
}

const (
	baseURL = "https://api.themoviedb.org/3/"
	defaultSearchParams = "?include_adult=false&page=1&query="
)

var (
	NoName ActorResource = ActorResource{Name: "NoName", Id: 0}
	NoTitle MovieResource = MovieResource{Title: "NoTitle", Id: 0}
)

func New(header string, timeout time.Duration) Client {
	return Client{
		httpClient: http.Client{
			Timeout: timeout,
		},
		cache: tmdbcache.New(),
		authHeader: header,
		searchFactor: 40,
		maxRoutines: 20,
	}
}

func (c *Client) SetSearchFactor(s int) {
	c.searchFactor = s
}

func (c *Client) SetMaxRoutines(r int) {
	c.maxRoutines = r
}

func (c *Client) GetMovies(actorId int) (map[int]struct{}, error) {
	if movies, ok := c.cache.GetMovies(actorId); ok {
		return movies, nil
	}

	url := baseURL
	//url += "discover/movie?include_adult=false&include_video=false&language=en-US&page=1&sort_by=popularity.desc&with_people="
	url += "person/"
	url += strconv.Itoa(actorId)
	url += "/movie_credits?language=en-US"
	res, err := getResource[Credits](url, c)
	if err != nil { return nil, err }

	movies := make(map[int]struct{})

	for _, movieRes := range res.Cast {
		if movieRes.Character != "" {
			movies[movieRes.Id] = struct{}{}
		}
	}
	/*for i := 0; i < min(len(res.Results), c.searchFactor); i++ {
		movie := res.Results[i]
		movies[movie.Id] = struct{}{}
	}*/

	c.cache.AddMovies(actorId, movies)
	return movies, nil
}

func (c *Client) GetActors(movieId int) (map[int]struct{}, error) {
	if actors, ok := c.cache.GetActors(movieId); ok {
		return actors, nil
	}

	url := baseURL + "movie/" + strconv.Itoa(movieId) + "/credits"
	res, err := getResource[Credits](url, c)
	if err != nil { return nil, err }

	actors := make(map[int]struct{})
	limit := c.searchFactor
	for _, actorRes := range res.Cast {
		if limit == 0 {
			break
		}
		if actorRes.Character != "" {
			actors[actorRes.Id] = struct{}{}
			limit--
		}
	}

	c.cache.AddActors(movieId, actors)
	return actors, nil
}

func (c *Client) GetMovieFromTitle(movieTitle string) (MovieResource, error) {
	query := fixStringForURL(movieTitle)
	url := baseURL + "search/movie" + defaultSearchParams + query
	
	res, err := getResource[MovieQueryResult](url, c)
	if err != nil { return MovieResource{}, err }

	if res.TotalResults > 0 {
		return res.Results[0], nil
	}
	fmt.Println(url)
	return NoTitle, nil
}

func (c *Client) GetActorFromName(actorName string) (ActorResource, error) {
	query := fixStringForURL(actorName)
	url := baseURL + "search/person" + defaultSearchParams + query

	res, err := getResource[ActorQueryResult](url, c)
	if err != nil { return ActorResource{}, err }

	if len(res.Results) > 0 {
		return res.Results[0], nil
	}
	return NoName, nil
}

func (c *Client) GetActorFromId(actorId int) (ActorResource, error) {
	url := baseURL + "person/" + strconv.Itoa(actorId)
	return getResource[ActorResource](url, c)
}

func (c *Client) GetMovieFromId(movieId int) (MovieResource, error) {
	url := baseURL + "movie/" + strconv.Itoa(movieId)
	return getResource[MovieResource](url, c)
}

func (c *Client) OverlappingActors(leftId, rightId int) ([]int, error) {
	url := baseURL + "movie/" + strconv.Itoa(leftId) + "/credits"
	creditResLeft, err := getResource[Credits](url, c)	
	if err != nil { return nil, err }
	url = baseURL + "movie/" + strconv.Itoa(rightId) + "/credits"
	creditResRight, err := getResource[Credits](url, c)
	if err != nil { return nil, err }
	
	actorsLeft := make(map[int]struct{})
	for _, actorRes := range creditResLeft.Cast {
		if actorRes.Character != "" {
			actorsLeft[actorRes.Id] = struct{}{}
		}
	}
	actorsRight := make(map[int]struct{})
	for _, actorRes := range creditResRight.Cast {
		if actorRes.Character != "" {
			actorsRight[actorRes.Id] = struct{}{}
		}
	}

	if len(actorsLeft) > len(actorsRight) {
		actorsLeft, actorsRight = actorsRight, actorsLeft
	}

	out := []int{}
	for actor := range actorsLeft {
		if _, ok := actorsRight[actor]; ok {
			out = append(out, actor)
		}
	}

	return out, nil
}

func (c *Client) newRequest(
	method, url string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil { return nil, err }

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", c.authHeader)
	return req, nil
}

func getResource[R resource](url string, c *Client) (R, error) {
	var zero R
	
	request, err := c.newRequest("GET", url, nil)
	if err != nil { return zero, err }

	response, err := c.httpClient.Do(request)
	if err != nil { return zero, err }
	defer response.Body.Close()

	dat, err := io.ReadAll(response.Body)
	if err != nil { return zero, err }

	var res R
	err = json.Unmarshal(dat, &res)
	if err != nil { return zero, err }

	return res, nil
}

func fixStringForURL(str string) string {
	return strings.ToLower(strings.ReplaceAll(str, " ", "+"))
}
