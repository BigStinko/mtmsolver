package tmdbcache

import (
	"sync"
)

type Cache struct {
	actorsMovies map[int]map[int]struct{}
	moviesActors map[int]map[int]struct{}
	neighbors    map[int]map[int]struct{}
	amu *sync.RWMutex
	mmu *sync.RWMutex
	nmu *sync.RWMutex
}

func New() Cache {
	return Cache{
		actorsMovies: make(map[int]map[int]struct{}),
		moviesActors: make(map[int]map[int]struct{}),
		neighbors:    make(map[int]map[int]struct{}),
		amu: &sync.RWMutex{},
		mmu: &sync.RWMutex{},
		nmu: &sync.RWMutex{},
	}
}

func Test(am, ma, n map[int]map[int]struct{}) Cache {
	return Cache{
		actorsMovies: am,
		moviesActors: ma,
		neighbors: n,
		amu: &sync.RWMutex{},
		mmu: &sync.RWMutex{},
		nmu: &sync.RWMutex{},
	}
}

func (c *Cache) GetMovies(actorId int) (map[int]struct{}, bool) {
	c.amu.RLock()
	defer c.amu.RUnlock()
	val, ok := c.actorsMovies[actorId]
	return val, ok
}

func (c *Cache) AddMovies(actorId int, movies map[int]struct{}) {
	c.amu.Lock()
	defer c.amu.Unlock()
	c.actorsMovies[actorId] = movies
}

func (c *Cache) AddMovie(actorId, movieId int) {
	c.amu.Lock()
	defer c.amu.Unlock()
	c.actorsMovies[actorId][movieId] = struct{}{}
}

func (c *Cache) GetActors(movieId int) (map[int]struct{}, bool) {
	c.mmu.RLock()
	defer c.mmu.RUnlock()
	val, ok := c.moviesActors[movieId]
	return val, ok
}

func (c *Cache) AddActors(movieId int, actors map[int]struct{}) {
	c.mmu.Lock()
	defer c.mmu.Unlock()
	c.moviesActors[movieId] = actors
}

func (c *Cache) AddActor(movieId, actorId int) {
	c.mmu.Lock()
	defer c.mmu.Unlock()
	c.moviesActors[movieId][actorId] = struct{}{}
}

func (c *Cache) GetNeighbors(movieId int) (map[int]struct{}, bool) {
	c.nmu.RLock()
	defer c.nmu.RUnlock()

	val, ok := c.neighbors[movieId]
	/*if !ok {
		val = make(map[int]struct{})
	}*/
	return val, ok
}

func (c *Cache) AddNeighbors(movieId int, neighbors map[int]struct{}) {
	c.nmu.Lock()
	defer c.nmu.Unlock()
	
	c.neighbors[movieId] = neighbors
}
