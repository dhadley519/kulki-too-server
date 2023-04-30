package game

type World map[int]map[int]*Position

func (w *World) getPosition(x int, y int) *Position {
	return (*w)[y][x]
}
