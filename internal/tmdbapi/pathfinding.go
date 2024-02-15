package tmdbapi

import (
	"container/list"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
)

var (
	ErrNoPath = errors.New("couldn't find path")
	ErrAlreadyVisited = errors.New("Already visited this node")
)

func GetPathExperimental(c *Client, src, dest string) ([]int, error) {
	c.SetSearchFactor(8)
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
	return c.runParallelSearchExperimental(srcRes.Id, destRes.Id)
}

func (c *Client) runParallelSearchExperimental(src, dest int) ([]int, error) {
	srcVisited, destVisited := sync.Map{}, sync.Map{}
	wg := sync.WaitGroup{}
	srcFoundCh, destFoundCh := make(chan int), make(chan int)
	PathCh := make(chan []int)
	errCh := make(chan error)
	doneCh := make(chan struct{})
	paths := make(map[int][][]int)
	defer close(srcFoundCh)
	defer close(destFoundCh)
	defer close(PathCh)
	defer close(errCh)
	defer close(doneCh)
	var err error
	go func() {
		for path := range PathCh {
			wg.Done()
			paths[path[0]] = append(paths[path[0]], path)
		}
	}()
	go func() {
		for range doneCh {
			wg.Done()
		}
	}()
	go func() {
		for e := range errCh {
			err = e
		}
	}()
	
	wg.Add(2)
	go c.parallelSearchExperimental(
		&srcVisited, &destVisited, &wg,
		src, srcFoundCh, destFoundCh,
		PathCh, doneCh, errCh,
	)
	go c.parallelSearchExperimental(
		&destVisited, &srcVisited, &wg,
		dest, destFoundCh, srcFoundCh,
		PathCh, doneCh, errCh,
	)
	wg.Wait()

	if err != nil { return nil, err }

	finalPath := []int{}
	
	for _, path := range paths {
		if len(path) != 2 {
			continue
		}
		path[0] = path[0][1:]
		slices.Reverse[[]int](path[0])
		path[0] = append(path[0], path[1]...)
		if len(path[0]) < len(finalPath) || len(finalPath) == 0 {
			finalPath = path[0]
		}
	}

	if len(finalPath) == 0 {
		return nil, ErrNoPath
	}

	return finalPath, nil
}

func (c *Client) parallelSearchExperimental(
	srcVisited, destVisited *sync.Map,
	pathWg *sync.WaitGroup,
	src int, iFoundCh chan<- int, theyFoundCh <-chan int,
	pathCh chan<- []int, doneCh chan<- struct{}, errChan chan<- error,
) {
	predecessors := sync.Map{}
	queueCh := make(chan int)
	defer close(queueCh)
	currentLevel := []int{src}
	nextLevel := []int{}
	found := newSafeBool()

	predecessors.Store(src, 0)
	srcVisited.Store(src, struct{}{})

	go func() {
		for node := range queueCh{
			nextLevel = append(nextLevel, node)
		}
	}()

	go func() {
		for node := range theyFoundCh {
			found.set()
			pathCh <- pathFromPredecessors(&predecessors, node)
		}
	}()

	for !found.isFound() {
		for len(currentLevel) > 0 && !found.isFound(){
			wg := sync.WaitGroup{}
			for i := 0; i < min(len(currentLevel), 10); i++ {
				current := currentLevel[0]
				currentLevel = currentLevel[1:]

				wg.Add(1)
				go func() {
					defer wg.Done()
					neighbors, err := c.GetNeighbors(current)
					if err != nil {
						errChan <- err
						found.set()
						return
					}
					node := c.evaluateNeighborsExperimental(
						srcVisited, destVisited, &predecessors,
						current,
						neighbors,
						queueCh,
					)
					if node != 0 {
						found.set()
						pathWg.Add(2)
						iFoundCh <- node
						pathCh <- pathFromPredecessors(&predecessors, node)
					}
				}()
			}
			wg.Wait()
		}
		currentLevel, nextLevel = nextLevel, currentLevel
	}
	doneCh <- struct{}{}
}

func GetPath(c *Client, src, dest string) ([]int, error) {
	c.SetSearchFactor(8)
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
	return c.runParallelSearch(srcRes.Id, destRes.Id)
}

// leftBfs takes two maps that will hold the visted sets of the searches
// starting from the "left", and "right" sides of the graph. When one side 
// of the search finds a node that is in the other sides visited set, that 
// means that the two searches have found eachother and they should have a 
// path to eachother through the node that they both have visited. The
// found channel is used to communicate to the other search the node that
// they need to return a path to.
func (c *Client) runParallelSearch(src, dest int) ([]int, error) {
	srcVisited, destVisited := sync.Map{}, sync.Map{}
	srcFoundCh, destFoundCh := make(chan int), make(chan int)
	srcPathch, destPathCh := make(chan []int), make(chan []int)
	errCh := make(chan error)
	srcFound, destFound, bothFound := false, false, false
	var srcPath []int
	var destPath []int

	go c.parallelSearch(
		&srcVisited, &destVisited,
		src, srcFoundCh, destFoundCh,
		srcPathch, errCh,
	)
	go c.parallelSearch(
		&destVisited, &srcVisited,
		dest, destFoundCh, srcFoundCh,
		destPathCh, errCh,
	)

	for !bothFound {
		select {
		case srcPath = <- srcPathch:
			srcFound = true
		case destPath = <- destPathCh:
			destFound = true
		case err := <- errCh:
			return nil, err
		default:
			if srcFound && destFound {
				bothFound = true
			}
		}
	}
	srcPath = srcPath[1:]
	slices.Reverse[[]int](srcPath)
	srcPath = append(srcPath, destPath...)
	return srcPath, nil
}


func (c *Client) parallelSearch(
	srcVisited, destVisited *sync.Map,
	src int, iFoundCh chan<- int, theyFoundCh <-chan int,
	pathCh chan<- []int, errChan chan<- error,
) {
	predecessors := sync.Map{}
	queueCh := make(chan int)
	w := newWorker()

	predecessors.Store(src, 0)
	srcVisited.Store(src, struct{}{})
	w.enqueue(src)

	go func() {
		for node := range queueCh{
			w.enqueue(node)
		}
	}()

	go func() {
		for node := range theyFoundCh {
			w.setFound()
			path := pathFromPredecessors(&predecessors, node)
			pathCh <- path
			break
		}
	}()

	for !w.isFound() {
		qLength := w.queue.Len()
		wg := sync.WaitGroup{}
		stopCh := make(chan struct{})

		for n := 0; n < min(qLength, 10); n++ {
			current := w.dequeue()
			wg.Add(1)

			go func() {  // TODO: test this
				defer wg.Done()
				neighbors, err := c.GetNeighbors(current)
				if err != nil {
					return
				}
				node := c.evaluateNeighbors(
					srcVisited, destVisited, &predecessors,
					current,
					neighbors,
					queueCh, stopCh,
				)
				if node != 0 {
					if !w.isFound() {  // TODO: check without this
						w.setFound()
						iFoundCh <- node
						path := pathFromPredecessors(&predecessors, node)
						pathCh <- path
					}
				}
			}()
		}
		wg.Wait()
	}
	close(queueCh)
}

func (c *Client) evaluateNeighbors(
	srcVisited, destVisited, predecessors *sync.Map,
	current int,
	neighbors map[int]struct{},
	queueCh chan<- int, stopCh <-chan struct{},
) int {
	for neighbor := range neighbors {
		select {
		case <- stopCh:
			return 0
		default:
			if _, ok := srcVisited.Load(neighbor); !ok {
				queueCh <- neighbor
				predecessors.Store(neighbor, current)
				srcVisited.Store(neighbor, struct{}{})
			}
			if _, ok := destVisited.Load(neighbor); ok {
				return neighbor
			}
		}
	}
	return 0
}

func (c *Client) evaluateNeighborsExperimental(
	srcVisited, destVisited, predecessors *sync.Map,
	current int,
	neighbors map[int]struct{},
	queueCh chan<- int,
) int {
	for neighbor := range neighbors {
		if _, ok := srcVisited.LoadOrStore(neighbor, struct{}{}); !ok {
			queueCh <- neighbor
			predecessors.Store(neighbor, current)
		}
		if _, ok := destVisited.Load(neighbor); ok {
			return neighbor
		}
	}
	return 0
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
		actor, err := c.getConnection(path[i - 1], p)
		if err != nil { return err }
		actorRes, err := c.GetActorFromId(actor)
		if err != nil { return err }
		fmt.Printf("Through: %s\nConnects to: %s\n", actorRes.Name, titles[i])
	}
	return nil
}


func (c *Client) GetNeighbors(movieId int) (map[int]struct{}, error) {
	neighbors := c.cache.GetNeighbors(movieId)

	actors, err := c.GetActors(movieId)
	if err != nil { return nil, err }

	for actor := range actors {
		movies, err := c.GetMovies(actor)
		if err != nil { return nil, err }
		if _, ok := movies[movieId]; !ok {
			movies[movieId] = struct{}{}
			c.cache.AddMovies(actor, movies)
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

type worker struct {
	queue *list.List
	found bool
	queueMux *sync.Mutex
	foundMux *sync.RWMutex
}

func newWorker() worker {
	return worker{
		queue: list.New(),
		found: false,
		queueMux: &sync.Mutex{},
		foundMux: &sync.RWMutex{},
	}
}

func (w *worker) isFound() bool {
	w.foundMux.RLock()
	defer w.foundMux.RUnlock()
	return w.found
}

func (w *worker) setFound() {
	w.foundMux.RLock()
	defer w.foundMux.RUnlock()
	w.found = true
}

func (w *worker) enqueue(node int) {
	w.queueMux.Lock()
	defer w.queueMux.Unlock()
	w.queue.PushBack(node)
}

func (w *worker) dequeue() int {
	w.queueMux.Lock()
	defer w.queueMux.Unlock()
	return w.queue.Remove(w.queue.Front()).(int)
}

type safeBool struct {
	value int32
}

func newSafeBool() safeBool {
	return safeBool{0}
}

func (s *safeBool) isFound() bool {
	val := atomic.LoadInt32(&s.value)
	return val != 0
}

func (s *safeBool) set() {
	atomic.StoreInt32(&s.value, 1)
}

func movieNotFoundError(title string) error {
	return errors.New(fmt.Sprintf("Could not find \"%s\"", title))
}
