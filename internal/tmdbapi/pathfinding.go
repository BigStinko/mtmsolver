package tmdbapi

import (
	"container/list"
	"errors"
	"slices"
)

func (c *Client) BFS(src, dest int) ([]int, error) {
	visited := make(map[int]struct{})
	distances := make(map[int]int)
	predecessors := make(map[int]int)
	queue := list.New()

	visited[src] = struct{}{}
	distances[src] = 0
	predecessors[src] = 0
	queue.PushBack(src)

	for queue.Len() != 0 {
		current := queue.Remove(queue.Front()).(int)

		if current == dest {
			path := []int{current}
			for predecessors[current] != 0 {
				current = predecessors[current]
				path = append(path, current)
			}

			slices.Reverse[[]int](path)
			return path, nil
		}

		if distances[current] > 6 {
			break
		}

		neighbors, err := c.getNeighbors(current)
		if err != nil { return nil, err }

		for _, neighbor := range neighbors {
			if _, ok := visited[neighbor]; !ok {
				visited[neighbor] = struct{}{}
				queue.PushBack(neighbor)
				distances[neighbor] = distances[current] + 1
				predecessors[neighbor] = current
			}
		}
	}

	return nil, errors.New("Couldn't find a path")
}

func (c *Client) getConnection(left, right int) (string, error) {
	actorsLeft, err := c.GetActors(left)
	if err != nil { return "", err }

	actorsRight, err := c.GetActors(right)
	if err != nil { return "", err }

	if len(actorsLeft) > len(actorsRight) {
		actorsLeft, actorsRight = actorsRight, actorsLeft
	}

	for _, actor := range actorsLeft {
		if slices.Contains[[]int](actorsRight, actor) {
			actorRes, err := c.GetActorFromId(actor)
			if err != nil { return "", err }
			return actorRes.Name, nil
		}
	}
	return "", errors.New("Failure finding neighbor connection")
}

func (c *Client) getNeighbors(movieId int) ([]int, error) {
	if neighbors, ok := c.cache.GetNeighbors(movieId); ok {
		return neighbors, nil
	}
	actors, err := c.GetActors(movieId)
	if err != nil { return nil, err }

	neighborsSet := make(map[int]struct{})

	for _, actor := range actors {
		movies, err := c.GetMovies(actor)
		if err != nil { return nil, err }

		for _, movie := range movies {
			if _, ok := neighborsSet[movie]; !ok {
				neighborsSet[movie] = struct{}{}
			}
		}
	}

	neighbors := make([]int, 0, len(neighborsSet))
	for neighbor := range neighborsSet {
		neighbors = append(neighbors, neighbor)
	}

	c.cache.AddNeighbors(movieId, neighbors)
	return neighbors, nil
}
