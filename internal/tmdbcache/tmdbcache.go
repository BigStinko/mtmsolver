package tmdbcache

import "sync"

type Cache struct {
	actorsMovies map[int][]int
	moviesActors map[int][]int
	neighbors map[int][]int
	amu *sync.RWMutex
	mmu *sync.RWMutex
	nmu *sync.RWMutex
}

func (c *Cache) GetMovies(actorId int) ([]int, bool) {
	c.amu.RLock()
	defer c.amu.RUnlock()
	val, ok := c.actorsMovies[actorId]
	return val, ok
}

func (c *Cache) GetActors(movieId int) ([]int, bool) {
	c.mmu.RLock()
	defer c.mmu.RUnlock()
	val, ok := c.moviesActors[movieId]
	return val, ok
}

func (c *Cache) AddMovies(actorId int, movies []int) {
	c.amu.Lock()
	defer c.amu.Unlock()

	c.actorsMovies[actorId] = movies
}

func (c *Cache) AddActors(movieId int, actors []int) {
	c.mmu.Lock()
	defer c.mmu.Unlock()
	
	c.moviesActors[movieId] = actors
}

func (c *Cache) GetNeighbors(movieId int) ([]int, bool) {
	c.nmu.RLock()
	defer c.nmu.RUnlock()

	val, ok := c.neighbors[movieId]
	return val, ok
}

func (c *Cache) AddNeighbors(movieId int, neighbors[]int) {
	c.nmu.Lock()
	defer c.nmu.Unlock()
	
	c.neighbors[movieId] = neighbors
}
