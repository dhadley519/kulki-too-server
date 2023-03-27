package game

import (
	"encoding/json"
	"github.com/beefsack/go-astar"
)

type Position struct {
	Point
	C *Color `json:"color"`
	W *World `json:"-"`
}

func (p *Position) MarshalJSON() ([]byte, error) {
	var c int
	if p.C != nil {
		c = p.C.value
	} else {
		c = 0
	}
	type k struct {
		X     int `json:"x"`
		Y     int `json:"y"`
		Color int `json:"color"`
	}
	return json.Marshal(k{p.X, p.Y, c})
}

func (p *Position) PathNeighbors() []astar.Pather {
	var neighbors []astar.Pather
	var offsets = [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	}
	for _, row := range offsets {
		x := p.X + row[0]
		y := p.Y + row[1]
		neighbor := p.W.getPosition(x, y)
		if neighbor != nil && neighbor.C == nil {
			neighbors = append(neighbors, neighbor)
		}
	}
	return neighbors
}

func (p *Position) PathNeighborCost(top astar.Pather) float64 {
	to := top.(*Position)
	c := float64(manhattan(p, to))
	return c
}

func (p *Position) PathEstimatedCost(top astar.Pather) float64 {
	to := top.(*Position)
	c := float64(manhattan(p, to))
	return c
}
