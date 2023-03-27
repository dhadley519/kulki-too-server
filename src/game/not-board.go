package game

import (
	"strconv"
	"strings"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var Colors = initColors()

func initColors() map[int]Color {
	var C = make(map[int]Color)
	C[1] = Color{"LightBlue", 1}
	C[2] = Color{"Blue", 2}
	C[3] = Color{"BlueViolet", 3}
	C[4] = Color{"Red", 4}
	C[5] = Color{"GreenYellow", 5}
	C[6] = Color{"Yellow", 6}
	C[7] = Color{"Pink", 7}
	C[8] = Color{"Orange", 8}
	C[9] = Color{"MediumSeaGreen", 9}
	return C
}

type direction int

const (
	NORTH direction = iota
	SOUTH
	WEST
	EAST
	NORTH_WEST
	SOUTH_WEST
	SOUTH_EAST
	NORTH_EAST
)

var OppositionalDirections = [][]direction{
	{NORTH, SOUTH},
	{WEST, EAST},
	{NORTH_WEST, SOUTH_EAST},
	{SOUTH_WEST, NORTH_EAST}}

type MovePath struct {
	From *Point `json:"from"`
	To   *Point `json:"to"`
}

type CommandType string

type BoardCommand interface {
	GetCommandType() CommandType
}

const (
	ADD       CommandType = "ADD"
	REMOVE    CommandType = "REMOVE"
	GAME_OVER CommandType = "GAME_OVER"
)

type Color struct {
	name  string
	value int
}

func ReviveBoard(width int,
	depth int,
	numberOfColors int,
	b string,
	nextBallColorsMemento string,
	score int,
) (*Board, []*BoardCommand) {
	var world = make(World)
	var commands []*BoardCommand
	//initialize empty game
	for i := 0; i < depth; i++ {
		world[i] = make(map[int]*Position)
	}
	unquote, err := strconv.Unquote(b)
	if err == nil {
		b = unquote
	}
	tiles := strings.Split(b, ``)
	for i := 0; i < depth; i++ {
		for j := 0; j < width; j++ {
			var index = (i * depth) + j
			d, err := strconv.Atoi(tiles[index])
			if err == nil {
				var pbc *Color
				if d > 0 {
					bc := Colors[d]
					pbc = &bc
				} else {
					pbc = nil
				}
				p := &Position{Point: Point{X: j, Y: i}, C: pbc, W: &world}
				world[i][j] = p
				if pbc != nil {
					var boardCommand BoardCommand = &AddCommand{Point: &Point{X: j, Y: i}, Color: pbc.value}
					commands = append(commands, &boardCommand)
				}
			}
		}
	}
	var bx = Board{world: &world,
		NumberOfColors: numberOfColors,
		Score:          score,
	}
	bx.SetNextBallColors(nextBallColorsMemento)
	return &bx, commands
}

func NewBoard(width int, depth int, numberOfColors int) *Board {
	var world = make(World)
	//initialize empty game
	for i := 0; i < depth; i++ {
		row := make(map[int]*Position)
		for j := 0; j < width; j++ {
			p := &Position{Point: Point{X: j, Y: i}, C: nil, W: &world}
			row[j] = p
		}
		(world)[i] = row
	}
	b := Board{world: &world,
		NumberOfColors: numberOfColors,
		nextBallColors: [3]int{0, 0, 0},
		Score:          0,
	}
	return &b
}

func abs(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}

func manhattan(start *Position, goal *Position) float32 {
	return float32(abs(start.X-goal.X) + abs(start.Y-goal.Y))
}
