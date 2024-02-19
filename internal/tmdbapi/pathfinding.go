package tmdbapi

import (
	"errors"
	"fmt"
	"slices"
	"sync"
)

var (
	ErrNoPath = errors.New("couldn't find path")
	ErrAlreadyVisited = errors.New("Already visited this node")
)

func GetPath(c *Client, src, dest string) ([]int, error) {
	//fmt.Printf("Finding path from: %s\nTo: %s\n", src, dest)
	srcRes, err := c.GetMovieFromTitle(src)
	if err != nil { return nil, err }
	if srcRes == NoTitle {
		return nil, movieNotFoundError(src)
	}
	destRes, err := c.GetMovieFromTitle(dest)
	if err != nil { return nil, err }
	if destRes == NoTitle {
		return nil, movieNotFoundError(dest)
	}
	if destRes.Id == srcRes.Id {
		return []int{srcRes.Id, srcRes.Id}, nil
	}
	path, err := c.runParallelSearch(srcRes.Id, destRes.Id)
	if err != nil {
		return nil, err
	}

	c.PrintPath(path)

	return path, nil
}

func (c *Client) runParallelSearch(src, dest int) ([]int, error) {
	srcCurrentLevel, destCurrentLevel := []int{src}, []int{dest}
	srcNextLevel, destNextLevel := []int{}, []int{}
	found := []int{}
	var err error
	srcVisited, destVisited := sync.Map{}, sync.Map{}
	srcPredecessors, destPredecessors := sync.Map{}, sync.Map{}
	srcVisited.Store(src, struct{}{})
	destVisited.Store(dest, struct{}{})
	srcPredecessors.Store(src, 0)
	destPredecessors.Store(dest, 0)

	for {
		srcNextLevel, srcCurrentLevel, found, err = c.getNextLevel(
			srcCurrentLevel, srcNextLevel,
			&srcVisited, &destVisited, &srcPredecessors,
		)
		if err != nil { return nil, err }
		if len(found) > 0 {
			break
		}
		destNextLevel, destCurrentLevel, found, err = c.getNextLevel(
			destCurrentLevel, destNextLevel,
			&destVisited, &srcVisited, &destPredecessors,
		)
		if err != nil { return nil, err }		
		if len(found) > 0 {
			break
		}
	}

	finalPath := []int{}

	for _, node := range found {
		srcPath := pathFromPredecessors(&srcPredecessors, node)
		destPath := pathFromPredecessors(&destPredecessors, node)
		srcPath = srcPath[1:]
		slices.Reverse[[]int](srcPath)
		srcPath = append(srcPath, destPath...)
		if len(srcPath) < len(finalPath) || len(finalPath) == 0 {
			if !slices.Contains[[]int](srcPath, src) || !slices.Contains[[]int](srcPath, dest) {
				continue
			}
			finalPath = srcPath
		}
	}

	return finalPath, nil
}

func (c *Client) getNextLevel(
	currentLevel, nextLevel []int,
	srcVisited, destVisited, predecessors *sync.Map,
) (cLevel[]int, nLevel[]int, found []int, finalErr error) {
	errCh := make(chan error)
	foundCh := make(chan int)
	queueCh := make(chan int)
	defer close(errCh)
	defer close(foundCh)
	defer close(queueCh)

	go func() {
		for e := range errCh {
			finalErr = e
		}
	}()
	go func() {
		for node := range foundCh {
			found = append(found, node)
		}
	}()
	go func() {
		for node := range queueCh {
			nextLevel = append(nextLevel, node)
		}
	}()

	for len(found) == 0 && len(currentLevel) > 0 {
		nextGroupSize := min(len(currentLevel), c.maxRoutines)
		searchGroup := currentLevel[:nextGroupSize]
		currentLevel = currentLevel[nextGroupSize:]
		wg := sync.WaitGroup{}
		
		for i := range searchGroup {
			current := searchGroup[i]
			wg.Add(1)
			go c.visitNeighbors(
				&wg,
				current,
				errCh,
				foundCh, queueCh,
				srcVisited, destVisited, predecessors,
			)
		}
		wg.Wait()
	}
	return currentLevel, nextLevel, found, finalErr
}

func (c *Client) visitNeighbors(
	wg *sync.WaitGroup,
	current int,
	errCh chan<- error,
	foundCh, queueCh chan<- int,
	srcVisited, destVisited, predecessors *sync.Map,
) {
	defer wg.Done()
	neighbors, err := c.GetNeighbors(current)
	if err != nil {
		errCh <- err
		foundCh <- -1
		return
	}
	for neighbor := range neighbors {
		if _, ok := srcVisited.LoadOrStore(neighbor, struct{}{}); !ok {
			predecessors.Store(neighbor, current)
			queueCh <- neighbor
		}
		if _, ok := destVisited.Load(neighbor); ok {
			foundCh <- neighbor
		}
	}
}

func (c *Client) GetNeighbors(movieId int) (map[int]struct{}, error) {
	neighbors, ok := c.cache.GetNeighbors(movieId)
	if ok {
		return neighbors, nil
	}

	actors, err := c.GetActors(movieId)
	if err != nil { return nil, err }

	for actor := range actors {
		movies, err := c.GetMovies(actor)
		if err != nil { return nil, err }

		for movie := range movies {
			if _, ok := neighbors[movie]; !ok {
				neighbors[movie] = struct{}{}
			}
		}
	}

	c.cache.AddNeighbors(movieId, neighbors)
	return neighbors, nil
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
		if _, ok := actorsRight[actor]; ok {
			return actor, nil
		}
	}

	return 0, errors.New("Failure finding neighbor connection")
}

func (c *Client) PrintPath(path []int) error {
	titles := make([]string, len(path))
	for i, p := range path {
		movieRes, err := c.GetMovieFromId(p)
		if err != nil { return err }
		titles[i] = movieRes.Title
	}
	fmt.Printf("Starting from: %s\n", titles[0])

	for i, p := range path {
		if i == 0 {
			continue
		}
		fmt.Printf("Through: ")
		actors, err := c.OverlappingActors(path[i - 1], p)
		if err != nil { return err }
		for _, actor := range actors {
			actorRes, err := c.GetActorFromId(actor)
			if err != nil { return err }
			fmt.Printf("%s,", actorRes.Name)			
		}
		fmt.Printf("\nConnects to: %s\n", titles[i])
	}
	return nil
}

func pathFromPredecessors(predecessors *sync.Map, src int) []int {
	path := []int{src}
	for {
		next, ok := predecessors.Load(src)
		if !ok || next == 0 {
			break
		}
		src = next.(int)
		path = append(path, src)
	}
	return path
}

func movieNotFoundError(title string) error {
	return errors.New(fmt.Sprintf("Could not find \"%s\"", title))
}
