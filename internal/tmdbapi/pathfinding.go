package tmdbapi

import (
	"container/list"
	"errors"
	"fmt"
	"slices"
	"sync"
)

var (
	ErrNoPath = errors.New("couldn't find path")
	ErrAlreadyVisited = errors.New("Already visited this node")
)

func (c *Client) GetPath(src, dest string) (int, error) {
	c.SetSearchFactor(7)
	//fmt.Printf("Finding path from: %s\nTo: %s\n", src, dest)
	srcRes, err := c.GetMovieFromTitle(src)
	if err != nil { return 0, err }
	destRes, err := c.GetMovieFromTitle(dest)
	if err != nil { return 0, err }
	path, err := c.runParallelSearch(srcRes.Id, destRes.Id)
	if err != nil { return 0, err }
	fmt.Println(path)

	if err != nil { return 0, err }
	if len(path) == 1 {
		fmt.Print("Start and destination are the same film")
		return 0, nil
	}

	/*out := ""
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
	}*/
	return len(path), nil
}

func (c *Client) runParallelSearch(src, dest int) ([]int, error) {
	var wg sync.WaitGroup
	var leftPath []int
	var rightPath []int
	leftMap, rightMap := sync.Map{}, sync.Map{}
	found := make(chan int, 1)
	var err error = nil
	wg.Add(2)
	go func() {
		defer wg.Done()
		var e error
		leftPath, e = c.parallelSearch(&leftMap, &rightMap, found, src, dest)
		if e != nil { err = e }
	}()
	go func () {
		defer wg.Done()
		var e error
		rightPath, e = c.parallelSearch(&rightMap, &leftMap, found, dest, src)
		if e != nil { err = e }
	}()

	wg.Wait()
	close(found)

	if err != nil {
		fmt.Printf("errored paths src: %v dest: %v\n", leftPath, rightPath)
		return nil, err
	}

	if len(rightPath) > 1 {
		rightPath = rightPath[1:]
	}
	slices.Reverse[[]int](leftPath)

	leftPath = append(leftPath, rightPath...)

	return leftPath, nil
}

// leftBfs takes two maps that will hold the visted sets of the searches
// starting from the "left", and "right" sides of the graph. When one side 
// of the search finds a node that is in the other sides visited set, that 
// means that the two searches have found eachother and they should have a 
// path to eachother through the node that they both have visited. The
// found channel is used to communicate to the other search the node that
// they need to return a path to.
func (c *Client) parallelSearch(
	leftVisited, rightVisited *sync.Map,
	found chan int,
	src, dest int,
) ([]int, error) {
	predecessors := map[int]int{src: 0}
	leftVisited.Store(src, struct{}{})
	queue := list.New()
	queue.PushBack(src)

	for queue.Len() != 0 {
		select {
		case f := <- found:
			if f == 0 { // indicates that the other search has errored, and
						// therefore this search can return without doing
						// anything
				return nil, nil
			}
			return pathFromPredecessors(predecessors, f), nil
		default:
			current := queue.Remove(queue.Front()).(int)

			if _, ok := rightVisited.Load(current); ok || current == dest {
				//fmt.Println("here")
				if len(found) < cap(found) {
					found <- current
				} else {
					queue.PushBack(current)
					continue
				}
				//fmt.Println("after")
				return pathFromPredecessors(predecessors, current), nil
			}
			
			neighbors, err := c.GetNeighbors(current)
			if err != nil { 
				found <- 0
				return nil, err 
			}
			
			for neighbor := range neighbors {
				if _, ok := leftVisited.LoadOrStore(neighbor, struct{}{}); !ok {
					queue.PushBack(neighbor)
					predecessors[neighbor] = current
				}
			}
		}
	}
	return nil, ErrNoPath
}

func (c *Client) parallelSearch3(
	leftVisited, rightVisited *sync.Map,
	src, dest int,
	pathChan []int, errChan error,
) {
	predecessors := sync.Map{}
	leftVisited.Store(src, struct{}{})
	queue := list.New()
	queue.PushBack(src)
	found := false
	queueCh := make(chan int)
	
	go func() {
		for node := range queueCh {
			queue.PushBack(node)
		}
	}()

	for queue.Len() != 0 && !found {
		currentLevel := make([]int, queue.Len())
		for i := range currentLevel {
			currentLevel[i] = queue.Remove(queue.Front()).(int)
			if _, ok := leftVisited.Load(currentLevel[i]); ok {
				return
			}
		}
		var wg sync.WaitGroup
		stopCh := make(chan struct{})
		errCh := make(chan error)
		wg.Add(len(currentLevel))

		for _, current := range currentLevel {
			go func() {
				defer wg.Done()
				neighbors, err := c.GetNeighbors(current)
				if err != nil {
					close(stopCh)
					errCh <- err
					return
				}
				for neighbor := range neighbors {
					select {
					case <- stopCh:
						return
					default:
						if _, ok := leftVisited.LoadOrStore(neighbor, struct{}{}); !ok {
							queueCh <- neighbor
							predecessors.Store(neighbor, current)
						}
					}
				}
				
			}()
		}

		wg.Wait()
	}
}

func (c *Client) processNeighbors() {
}

func (c *Client) GetNeighbors(movieId int) (map[int]struct{}, error) {
	var neighbors map[int]struct{}
	neighbors, ok := c.cache.GetNeighbors(movieId)
	if !ok { neighbors = make(map[int]struct{}) }

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
	for next, ok := predecessors.Load(src); next != 0 && ok; {
		src = next.(int)
		path = append(path, src)
	}
	return path
}

func (c *Client) parallelSearch2(
	leftVisited, rightVisited *sync.Map,
	src, dest int,
	foundCh chan int, pathCh chan<- []int, errChan chan<- error,
) {
	predecessors := sync.Map{}
	queue := list.New()
	queueCh := make(chan int)
	found := false
	errored := false
	connectingNode := 0

	predecessors.Store(src, 0)
	leftVisited.Store(src, struct{}{})
	queue.PushBack(src)

	go func() {
		for node := range queueCh{
			queue.PushBack(node)
		}
	}()

	for !found {
		qLength := queue.Len()
		wg := sync.WaitGroup{}
		closeCh := make(chan struct{})

		for n := 0; n < qLength; n++ {
			currentNode := queue.Remove(queue.Front()).(int)
			wg.Add(1)

			go func(current int) {  // TODO: test this
				defer wg.Done()
				neighbors, err := c.GetNeighbors(current)
				if err != nil {
					close(closeCh)
					found = true
					errChan <- err
					return
				}
				node := c.evaluateNeighbors(
					leftVisited, rightVisited, &predecessors,
					current,
					neighbors,
					queueCh, closeCh,
				)
				if node != 0 {
					connectingNode = node
					found = true
					close(closeCh)
				}
			}(currentNode)
		}
		wg.Wait()
		close(closeCh)
	}
	close(queueCh)
	if connectingNode == 0 || errored {
		return
	}
	//TODO: if found
}

func (c *Client) evaluateNeighbors(
	leftVisited, rightVisited, predecessors *sync.Map,
	current int,
	neighbors map[int]struct{},
	queue chan<- int, closeCh <-chan struct{},
) int {
	for neighbor := range neighbors {
		select {
		case <- closeCh:
			return 0
		default:
			if _, ok := rightVisited.Load(neighbor); ok {
				return neighbor
			}
			if _, ok := leftVisited.LoadOrStore(neighbor, struct{}{}); !ok {
				queue <- neighbor
				predecessors.Store(neighbor, current)
			}
		}
	}
	return 0
}
