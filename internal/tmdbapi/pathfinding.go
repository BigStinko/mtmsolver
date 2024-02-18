package tmdbapi

import (
	"container/list"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/BigStinko/mtmsolver/internal/priorityqueue"
)

var (
	ErrNoPath = errors.New("couldn't find path")
	ErrAlreadyVisited = errors.New("Already visited this node")
)

func GetPath3(c *Client, src, dest string) ([]int, error) {
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
	path, err := c.runParallelSearch3(srcRes.Id, destRes.Id)
	if err != nil {
		if err == ErrNoPath {
			return GetPath3(c, src, dest)
		}
		return nil, err
	}
	return path, nil
}

func (c *Client) runParallelSearch3(src, dest int) ([]int, error) {
	srcVisited, destVisited := sync.Map{}, sync.Map{}
	wg := sync.WaitGroup{}
	srcFoundCh, destFoundCh := make(chan int), make(chan int)
	srcPathCh, destPathCh := make(chan []int), make(chan []int)
	errCh := make(chan error)
	srcPaths, destPaths := make(map[int][]int), make(map[int][]int)
	defer close(srcFoundCh)
	defer close(destFoundCh)
	defer close(srcPathCh)
	defer close(destPathCh)
	defer close(errCh)
	var err error

	go func() {
		for path := range srcPathCh {
			wg.Done()
			srcPaths[path[0]] = path
		}
	}()
	go func() {
		for path := range destPathCh {
			wg.Done()
			destPaths[path[0]] = path
		}
	}()
	go func() {
		for e := range errCh {
			wg.Done()
			err = e
		}
	}()
	wg.Add(2)
	go c.parallelSearch3(
		&srcVisited, &destVisited,
		&wg,
		srcFoundCh,
		destFoundCh,
		srcPathCh,
		errCh,
		src,
	)
	go c.parallelSearch3(
		&destVisited, &srcVisited,
		&wg,
		destFoundCh,
		srcFoundCh,
		destPathCh,
		errCh,
		dest,
	)

	wg.Wait()

	if err != nil {
		return nil, err
	}

	finalPath := []int{}

	for con, sPath := range srcPaths {
		dPath := destPaths[con]
		if len(sPath) < 1 || len(dPath) < 1 {
			continue
		}
		sPath = sPath[1:]
		slices.Reverse[[]int](sPath)
		sPath = append(sPath, dPath...)
		if len(sPath) < len(finalPath) || len(finalPath) == 0 {
			finalPath = sPath
		}
	}
	if len(finalPath) == 0 {
		return nil, ErrNoPath
	}
	if !slices.Contains[[]int](finalPath, src) || !slices.Contains[[]int](finalPath, dest) {
		return nil, ErrNoPath
	}

	return finalPath, nil
}

func (c *Client) parallelSearch3(
	iVisited, theyVisited *sync.Map,
	callerWg *sync.WaitGroup,
	iFound chan<- int,
	theyFound <-chan int,
	pathCh chan<- []int,
	errCh chan<- error,
	src int,
) {
	defer callerWg.Done()
	predecessors := sync.Map{}
	currentLevel, nextLevel := []int{src}, []int{}
	found := newSafeBool()
	wg := sync.WaitGroup{}
	queueCh := make(chan int)

	iVisited.Store(src, struct{}{})
	predecessors.Store(src, 0)

	go func() {
		for node := range queueCh {
			nextLevel = append(nextLevel, node)
		}
	}()

	go func() {
		for node := range theyFound {
			found.set()
			pathCh <- pathFromPredecessors(&predecessors, node)
		}
	}()

	for !found.isFound() {
		nextGroupSize := min(len(currentLevel), c.maxRoutines)
		searchGroup := currentLevel[:nextGroupSize]
		currentLevel = currentLevel[nextGroupSize:]
		
		for i := range searchGroup {
			current := searchGroup[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				neighbors, err := c.GetNeighbors(current)
				if err != nil {
					found.set()
					callerWg.Add(1)
					errCh <- err
					return
				}
				node := evaluateNeighbors3(
					iVisited, theyVisited, &predecessors,
					queueCh,
					neighbors,
					current,
				)
				if node != 0 && !found.isFound(){
					callerWg.Add(2) // TODO: test this
					found.set()
					iFound <- node
					pathCh <- pathFromPredecessors(&predecessors, node)
				}
			}()
		}
		wg.Wait()
		if len(currentLevel) == 0 {
			currentLevel, nextLevel = nextLevel, currentLevel
		}
	}
}

func evaluateNeighbors3(
	iVisited, theyVisited, predecessors *sync.Map,
	queueCh chan<- int,
	neighbors map[int]struct{},
	current int,
) int {
	for neighbor := range neighbors {
		if _, ok := iVisited.LoadOrStore(neighbor, struct{}{}); !ok {
			predecessors.Store(neighbor, current)
			queueCh <- neighbor
		}
		if _, ok := theyVisited.Load(neighbor); ok {
			return neighbor
		}
	}
	return 0
}

func GetPath2(c *Client, src, dest string) ([]int, error) {
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
	path, err := c.runParallelSearch2(srcRes.Id, destRes.Id)
	if err != nil {
		if err == ErrNoPath {
			return GetPath2(c, src, dest)
		}
		return nil, err
	}
	return path, nil
}

// leftBfs takes two maps that will hold the visted sets of the searches
// starting from the "left", and "right" sides of the graph. When one side 
// of the search finds a node that is in the other sides visited set, that 
// means that the two searches have found eachother and they should have a 
// path to eachother through the node that they both have visited. The
// found channel is used to communicate to the other search the node that
// they need to return a path to.
func (c *Client) runParallelSearch2(src, dest int) ([]int, error) {
	srcVisited, destVisited := sync.Map{}, sync.Map{}
	srcFoundCh, destFoundCh := make(chan int), make(chan int)
	srcPathCh, destPathCh := make(chan []int), make(chan []int)
	errCh := make(chan error)
	var err error
	wg := sync.WaitGroup{}
	srcPaths, destPaths := make(map[int][]int), make(map[int][]int)
	defer close(srcFoundCh)
	defer close(destFoundCh)
	defer close(srcPathCh)
	defer close(destPathCh)
	defer close(errCh)
	
	go func() {
		for path := range srcPathCh {
			srcPaths[path[0]] = path
			wg.Done()
		}
	}()

	go func() {
		for path := range destPathCh {
			destPaths[path[0]] = path
			wg.Done()
		}
	}()

	go func() {
		for e := range errCh {
			err = e
			wg.Done()
		}
	}()

	wg.Add(2)
	go c.parallelSearch2(
		&srcVisited, &destVisited,
		src, &wg, srcFoundCh, destFoundCh,
		srcPathCh, errCh,
	)
	go c.parallelSearch2(
		&destVisited, &srcVisited,
		dest, &wg, destFoundCh, srcFoundCh,
		destPathCh, errCh,
	)
	wg.Wait()
	if err != nil {
		return nil, err
	}

	finalPath := []int{}

	for con, sPath := range srcPaths {
		dPath := destPaths[con]
		if len(sPath) < 1 || len(dPath) < 1 {
			continue
		}
		sPath = sPath[1:]
		slices.Reverse[[]int](sPath)
		sPath = append(sPath, dPath...)
		if len(sPath) < len(finalPath) || len(finalPath) == 0 {
			finalPath = sPath
		}
	}
	if len(finalPath) == 0 {
		return nil, ErrNoPath
	}
	if !slices.Contains[[]int](finalPath, src) || !slices.Contains[[]int](finalPath, dest) {
		return nil, ErrNoPath
	}

	return finalPath, nil
}


func (c *Client) parallelSearch2(
	srcVisited, destVisited *sync.Map,
	src int, callerWg *sync.WaitGroup,
	iFoundCh chan<- int, theyFoundCh <-chan int,
	pathCh chan<- []int, errChan chan<- error,
) {
	defer callerWg.Done()
	predecessors := sync.Map{}
	distances := sync.Map{}
	queue := priorityqueue.NewSafePQ()
	found := newSafeBool()

	predecessors.Store(src, 0)
	srcVisited.Store(src, struct{}{})
	distances.Store(src, 0)
	queue.Push(src, 0)

	go func() {
		for node := range theyFoundCh {
			found.set()
			path := pathFromPredecessors(&predecessors, node)
			pathCh <- path
		}
	}()

	for !found.isFound() {
		wg := sync.WaitGroup{}

		for n := 0; n < min(queue.Len(), c.maxRoutines); n++ {
			current := queue.Pop()
			wg.Add(1)

			go func() {  // TODO: test this
				defer wg.Done()
				neighbors, err := c.GetNeighbors(current)
				if err != nil {
					wg.Add(1)
					errChan <- err
					return
				}
				node := c.evaluateNeighbors2(
					srcVisited, destVisited, &predecessors, &distances,
					&queue,
					current,
					neighbors,
				)
				if node != 0 {
					found.set()
					callerWg.Add(2)
					iFoundCh <- node
					path := pathFromPredecessors(&predecessors, node)
					pathCh <- path
				}
			}()
		}
		wg.Wait()
	}
}

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
	srcPathCh, destPathCh := make(chan []int), make(chan []int)
	errCh := make(chan error)
	srcFound, destFound, bothFound := false, false, false
	var srcPath []int
	var destPath []int

	go c.parallelSearch(
		&srcVisited, &destVisited,
		src, srcFoundCh, destFoundCh,
		srcPathCh, errCh,
	)
	go c.parallelSearch(
		&destVisited, &srcVisited,
		dest, destFoundCh, srcFoundCh,
		destPathCh, errCh,
	)

	for !bothFound {
		select {
		case srcPath = <- srcPathCh:
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

		for n := 0; n < min(qLength, c.maxRoutines); n++ {
			current := w.dequeue()
			wg.Add(1)

			go func() {  // TODO: test this
				defer wg.Done()
				neighbors, err := c.GetNeighbors(current)
				if err != nil {
					errChan <- err
					close(stopCh)
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

func (c *Client) evaluateNeighbors2(
	srcVisited, destVisited, predecessors, distances *sync.Map,
	queue *priorityqueue.SafePriorityQueue,
	current int,
	neighbors map[int]struct{},
) int {
	for neighbor := range neighbors {
		if _, ok := srcVisited.LoadOrStore(neighbor, struct{}{}); !ok {
			val, _ := distances.Load(current)
			distance := val.(int)
			distances.Store(neighbor, distance + 1)
			queue.Push(neighbor, distance + 1)
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
	neighbors, ok := c.cache.GetNeighbors(movieId)
	if ok {
		return neighbors, nil
	}

	actors, err := c.GetActors(movieId)
	if err != nil { return nil, err }

	for actor := range actors {
		if actor == 1658615 {
			fmt.Println("jeff pope spotted")
		}
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

type safeQueue struct {
	queue *list.List
	qMux *sync.Mutex
}

func newQueue() safeQueue {
	return safeQueue{
		queue: list.New(),
		qMux: &sync.Mutex{},
	}
}

func (q *safeQueue) enqueue(node int) {
	q.qMux.Lock()
	defer q.qMux.Unlock()
	q.queue.PushBack(node)
}

func (q *safeQueue) dequeue() int {
	q.qMux.Lock()
	defer q.qMux.Unlock()
	return q.queue.Remove(q.queue.Front()).(int)
}

func (q *safeQueue) Len() int {
	q.qMux.Lock()
	defer q.qMux.Unlock()
	return q.queue.Len()
}

func movieNotFoundError(title string) error {
	return errors.New(fmt.Sprintf("Could not find \"%s\"", title))
}


