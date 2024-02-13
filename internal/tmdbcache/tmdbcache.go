package tmdbcache

import (
	_"fmt"
	"sync"
)

type Cache struct {
	actorsMovies sync.Map
	moviesActors sync.Map
	neighbors    sync.Map
}

func New() Cache {
	return Cache{
		actorsMovies: sync.Map{},
		moviesActors: sync.Map{},
		neighbors:    sync.Map{},
	}
}

func (c *Cache) GetMovies(actorId int) (map[int]struct{}, bool) {
	//fmt.Println("GetMovies")
	val, ok := c.actorsMovies.Load(actorId)
	if !ok {
		return nil, ok
	}
	m, ok := val.(map[int]struct{})
	return m, ok
}

func (c *Cache) AddMovies(actorId int, movies map[int]struct{}) {
	//fmt.Println("AddMovies")
	c.actorsMovies.Store(actorId, movies)
}

func (c *Cache) AddMovie(actorId, movieId int) {
	//fmt.Println("AddMovie")
	val, _ := c.actorsMovies.Load(actorId)
	movies := val.(map[int]struct{}) 
	movies[movieId] = struct{}{}
	c.actorsMovies.Store(actorId, movies)
}

func (c *Cache) GetActors(movieId int) (map[int]struct{}, bool) {
	//fmt.Println("GetActors")
	val, ok := c.moviesActors.Load(movieId)
	if !ok {
		return nil, ok
	}
	a, ok := val.(map[int]struct{})
	return a, ok
}

func (c *Cache) AddActors(movieId int, actors map[int]struct{}) {
	//fmt.Println("AddActors")
	c.moviesActors.Store(movieId, actors)
}

func (c *Cache) AddActor(movieId, actorId int) {
	//fmt.Println("AddActors")
	val, _ := c.moviesActors.Load(movieId)
	actors := val.(map[int]struct{})
	actors[actorId] = struct{}{}
	c.moviesActors.Store(movieId, actors)
}

func (c *Cache) GetNeighbors(movieId int) map[int]struct{} {
	//fmt.Println("GetNeighbors")
	val, ok := c.neighbors.Load(movieId)
	if !ok {
		return make(map[int]struct{})
	}
	return val.(map[int]struct{})
}

func (c *Cache) AddNeighbors(movieId int, neighbors map[int]struct{}) {
	//fmt.Println("AddNeighbors")
	c.neighbors.Store(movieId, neighbors)
}
