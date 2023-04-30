package game

import "encoding/json"

type PointCommand interface {
	P()
}

type ColorCommand interface {
	C()
}

type AddCommand struct {
	Point *Point
	Color int
	BoardCommand
	PointCommand
	ColorCommand
}

func (a *AddCommand) GetCommandType() CommandType {
	return ADD
}

func (a *AddCommand) P() *Point {
	return a.Point
}

func (a *AddCommand) C() int {
	return a.Color
}

func (a *AddCommand) MarshalJSON() ([]byte, error) {
	type k struct {
		X       int         `json:"x"`
		Y       int         `json:"y"`
		Color   int         `json:"color"`
		Command CommandType `json:"command"`
	}
	return json.Marshal(k{a.Point.X, a.Point.Y, a.Color, a.GetCommandType()})
}

type GameOverCommand struct {
	BoardCommand
}

func (g *GameOverCommand) GetCommandType() CommandType {
	return GAME_OVER
}

func (g *GameOverCommand) MarshalJSON() ([]byte, error) {
	type k struct {
		Command CommandType `json:"command"`
	}
	return json.Marshal(k{g.GetCommandType()})
}
