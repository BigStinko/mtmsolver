package tmdbapi

import (
	"container/list"
	"errors"
	"fmt"
	"slices"
	"sync"
)

var ErrNoPath = errors.New("couldn't find path")

func (c *Client) GetPath(src, dest string) (string, error) {
	c.SetSearchFactor(7)
	fmt.Printf("Finding path from: %s\nTo: %s\n", src, dest)
	srcRes, err := c.GetMovieFromTitle(src)
	if err != nil { return "", err }
	destRes, err := c.GetMovieFromTitle(dest)
	if err != nil { return "", err }
	path, err := c.runLeftSearch(srcRes.Id, destRes.Id)
	if err != nil { return "", err }

	/*pathCh := make(chan []int)
	errCh := make(chan error)
	var wg sync.WaitGroup
	wg.Add(2)
	go c.runSearch(pathCh, errCh, &wg, srcRes.Id, destRes.Id)
	go c.runSearch(pathCh, errCh, &wg, destRes.Id, srcRes.Id)
	go func() {
		wg.Wait()
		close(pathCh)
		close(errCh)
	}()

	var path []int
	select {
	case path = <- pathCh:
	case err = <- errCh:
		return "", err
	}*/

	if err != nil { return "", err }
	if len(path) == 1 {
		fmt.Printf("Start and destination are the same film")
		return "", nil
	}

	out := ""
	if path[0] != srcRes.Id {
		slices.Reverse[[]int](path)
	}

	out += fmt.Sprintf("Starting from: %s\n", srcRes.Title)
	for i, p := range path {
		if p == srcRes.Id {
			continue
		}
		actor, err := c.getConnection(path[i - 1], p)
		if err != nil { return "", err }
		actorRes, err := c.GetActorFromId(actor)
		if err != nil { return "", err }
		movieRes, err := c.GetMovieFromId(p)
		if err != nil { return "", err }
		out += fmt.Sprintf("Through: %s\nConnects to: %s\n", actorRes.Name, movieRes.Title) 
	}
	return out, nil
}

func (c *Client) runSearch(
	ch chan<- []int, errCh chan<- error,
	wg *sync.WaitGroup,
	src, dest int) {
	defer wg.Done()
	result, err := c.bfs(src, dest)
	
	if err != nil {
		select {
		case errCh <- err:
		default:
		}
		return
	}
	select {
	case ch <- result:
	default:
	}
}

func (c *Client) bfs(src, dest int) ([]int, error) {
	visited := make(map[int]struct{})
	predecessors := make(map[int]int)
	distances := make(map[int]int)
	queue := list.New()
	visited[src] = struct{}{}
	predecessors[src] = 0
	distances[src] = 0
	queue.PushBack(src)

	for queue.Len() != 0 {
		current := queue.Remove(queue.Front()).(int)

		if current == dest {
			path := pathFromPredecessors(predecessors, current)
			slices.Reverse[[]int](path)
			return path, nil
		}

		neighbors, err := c.getNeighbors(current)
		if err != nil { return nil, err }

		for neighbor := range neighbors {
			if _, ok := visited[neighbor]; !ok {
				visited[neighbor] = struct{}{}
				queue.PushBack(neighbor)
				predecessors[neighbor] = current
				distances[neighbor] = distances[current] + 1
			}
		}
	}

	return nil, ErrNoPath
}

func (c *Client) runLeftSearch(src, dest int) ([]int, error) {
	var wg sync.WaitGroup
	var leftPath []int
	var rightPath []int
	leftMap, rightMap := sync.Map{}, sync.Map{}
	found := make(chan int)
	var err error = nil
	wg.Add(2)
	go func() {
		defer wg.Done()
		var e error
		leftPath, e = c.leftBfs(&leftMap, &rightMap, found, src)
		if e != nil { err = e }
	}()
	go func () {
		defer wg.Done()
		var e error
		rightPath, e = c.leftBfs(&rightMap, &leftMap, found, dest)
		if e != nil { err = e }
	}()

	wg.Wait()
	close(found)

	if err != nil {
		return nil, err
	}
	fmt.Println(leftPath)
	fmt.Println(rightPath)

	if len(rightPath) > 1 {
		rightPath = rightPath[1:]
	}
	slices.Reverse[[]int](leftPath)

	leftPath = append(leftPath, rightPath...)
	fmt.Println(leftPath)

	return leftPath, nil
}

func (c *Client) leftBfs(leftVisited, rightVisited *sync.Map, found chan int, src int) ([]int, error) {
	predecessors := map[int]int{src: 0}
	leftVisited.Store(src, struct{}{})
	queue := list.New()
	queue.PushBack(src)

	for queue.Len() != 0 {
		select {
		case f := <- found:
			if f == 0 {
				return nil, nil
			}
			return pathFromPredecessors(predecessors, f), nil
		default:
			current := queue.Remove(queue.Front()).(int)

			if _, ok := rightVisited.Load(current); ok {
				found <- current
				return pathFromPredecessors(predecessors, current), nil
			}
			
			neighbors, err := c.getNeighbors(current)
			if err != nil { 
				found <- 0
				return nil, err 
			}
			
			for neighbor := range neighbors {
				if _, ok := leftVisited.Load(neighbor); !ok {
					leftVisited.Store(neighbor, struct{}{})
					queue.PushBack(neighbor)
					predecessors[neighbor] = current
				}
			}
		}
	}
	return nil, ErrNoPath
}

func pathFromPredecessors(predecessors map[int]int, src int) []int {
	path := []int{src}
	for predecessors[src] != 0 {
		src = predecessors[src]
		path = append(path, src)
	}
	return path
}

func (c *Client) getConnection(left, right int) (int, error) {
	actorsLeft, err := c.GetActors(left)
	if err != nil { return 0, err }

	actorsRight, err := c.GetActors(right)
	if err != nil { return 0, err }

	if len(actorsLeft) > len(actorsRight) {
		actorsLeft, actorsRight = actorsRight, actorsLeft
	}

	for actor := range actorsLeft {
		movies, err := c.GetMovies(actor)
		if err != nil { return 0, err }

		for movie := range movies {
			if movie == right {
				return actor, nil
			}
		}
	}

	for actor := range actorsRight {
		movies, err := c.GetMovies(actor)
		if err != nil { return 0, err }

		for movie := range movies {
			if movie == left {
				return actor, nil
			}
		}
	}

	for actor := range actorsLeft {
		if _, ok := actorsRight[actor]; ok {
			return actor, nil
		}
	}

	return 0, errors.New("Failure finding neighbor connection")
}

func (c *Client) getNeighbors(movieId int) (map[int]struct{}, error) {
	if neighbors, ok := c.cache.GetNeighbors(movieId); ok {
		return neighbors, nil
	}
	actors, err := c.GetActors(movieId)
	if err != nil { return nil, err }

	neighbors := make(map[int]struct{})

	for actor := range actors {
		movies, err := c.GetMovies(actor)
		if err != nil { return nil, err }
		if _, ok := movies[movieId]; !ok {
			//fmt.Println("actor is in movie, but movie is not in actors movies")
			movies[movieId] = struct{}{}
			//c.cache.AddMovie(actor, movieId)
		}

		for movie := range movies {
			if _, ok := neighbors[movie]; !ok {
				neighbors[movie] = struct{}{}
			}
		}
	}

	c.cache.AddNeighbors(movieId, neighbors)
	return neighbors, nil
}

