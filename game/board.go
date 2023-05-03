package game

import (
	"github.com/beefsack/go-astar"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

type Board struct {
	world          *World
	NumberOfColors int
	nextBallColors [3]int
	Score          int
}

func (b *Board) Depth() int {
	return len(*b.world)
}

func (b *Board) Width() int {
	return len((*b.world)[0])
}

func (b *Board) randomNext3Colors() {
	for i := 0; i < 3; i++ {
		b.nextBallColors[i] = rand.Intn(b.NumberOfColors) + 1
	}
}

func (b *Board) addNext3Balls() []*BoardCommand {
	var commands []*BoardCommand
	if b.EmptyTileCount() < 3 {
		return commands
	}
	for i := 0; i < 3; {
		y := rand.Intn(b.Depth())
		x := rand.Intn(b.Width())
		var p = (*b.world)[y][x]
		if p.C == nil {
			var c = Colors[b.nextBallColors[i]]
			var command BoardCommand = &AddCommand{Point: &Point{X: p.X, Y: p.Y}, Color: c.value}
			p.C = &c
			commands = append(commands, &command)
			i++
		}
	}
	b.randomNext3Colors()
	return commands
}

func (b *Board) OnStartMessageReceipt() []*BoardCommand {
	b.randomNext3Colors()
	return b.addNext3Balls()
}

func (b *Board) OnMoveMessageReceipt(move *MovePath) ([]*BoardCommand, bool) {
	success := b.move(move)
	var changes []*BoardCommand
	if success {
		var position1 = b.world.getPosition(move.To.X, move.To.Y)
		var commandA BoardCommand = &AddCommand{Point: &Point{X: position1.X, Y: position1.Y}, Color: position1.C.value}
		changes = append(changes, &commandA)
		var position2 = b.world.getPosition(move.From.X, move.From.Y)
		var commandR BoardCommand = &RemoveCommand{Point: &Point{X: position2.X, Y: position2.Y}}
		changes = append(changes, &commandR)
		removals := b.getMatchRemovalsAndUpdateScore(b.world.getPosition(move.To.X, move.To.Y))
		if len(removals) > 0 {
			for _, r := range removals {
				var bc BoardCommand = r
				changes = append(changes, &bc)
			}
		} else {
			var adds []*BoardCommand = b.addNext3Balls()
			var failedToAdd3Balls bool = len(adds) < 3
			for _, a := range adds {
				changes = append(changes, a)
				var bc BoardCommand = *a
				if bc.GetCommandType() == GAME_OVER {
					changes = append(changes, a)
				}
				if bc.GetCommandType() == ADD {
					var a = bc.(*AddCommand)
					var x int = a.P().X
					var y int = a.P().Y
					removals := b.getMatchRemovalsAndUpdateScore(b.world.getPosition(x, y))
					for _, r := range removals {
						var bc BoardCommand = r
						changes = append(changes, &bc)
					}
				}
			}
			if failedToAdd3Balls || b.EmptyTileCount() == 0 {
				var gameOver BoardCommand
				gameOver = &GameOverCommand{}
				changes = append(changes, &gameOver)
			}
		}
	}
	return changes, success
}

func (b *Board) OnPathFindReceipt(path *MovePath) ([]*Position, bool) {
	solution, success := b.FindSolution(b.world.getPosition(path.From.X, path.From.Y), b.world.getPosition(path.To.X, path.To.Y))
	if success {
		return solution, success
	}
	return nil, false
}

func (b *Board) move(move *MovePath) bool {
	var from = b.world.getPosition(move.From.X, move.From.Y)
	var to = b.world.getPosition(move.To.X, move.To.Y)
	_, success := b.FindSolution(from, to)
	if success && from.C != nil {
		to.C = from.C
		b.release(from)
		return true
	}
	log.Println("failed To move")
	return false
}

func (b *Board) release(position *Position) bool {
	if position.C != nil {
		position.C = nil
		return true
	}
	log.Println("failed To release")
	return false
}

func (b *Board) getMatchRemovalsAndUpdateScore(position *Position) []*RemoveCommand {
	var matches []*Position
	for _, row := range OppositionalDirections {
		nSMatches := b.getMatchesBiDirectional(position, row[0], row[1])
		if len(nSMatches) >= 4 {
			matches = append(matches, nSMatches...)
		}
	}
	if len(matches) > 0 {
		matches = append(matches, position)
		var commands []*RemoveCommand
		for _, m := range matches {
			commands = append(commands, &RemoveCommand{Point: &Point{X: m.X, Y: m.Y}})
			b.release(m)
		}
		b.Score += len(matches) * len(matches)
		return commands
	}
	return []*RemoveCommand{}
}

func (b *Board) getMatchesBiDirectional(position *Position, d1 direction, d2 direction) []*Position {
	var matches []*Position
	matches = b.getMatchesUniDirectional(position, d1, matches)
	matches = b.getMatchesUniDirectional(position, d2, matches)
	return matches
}

func (b *Board) getMatchesUniDirectional(position *Position, d direction, matches []*Position) []*Position {
	neighbor := b.getNeighbor(position, d)
	if neighbor != nil && neighbor.C != nil && neighbor.C.value == position.C.value {
		matches = append(matches, neighbor)
		return b.getMatchesUniDirectional(neighbor, d, matches)
	}
	return matches
}

func (b *Board) FindSolution(start *Position, goal *Position) ([]*Position, bool) {

	pather, _, success := astar.Path(start, goal)

	var result []*Position

	for _, p := range pather {
		result = append(result, p.(*Position))
	}

	return result, success
}

func (b *Board) Memento() string {
	var val []string
	for i := 0; i < b.Depth(); i++ {
		row := (*b.world)[i]
		for j := 0; j < b.Width(); j++ {
			bc := row[j].C
			bci := 0
			if bc != nil {
				bci = bc.value
			}
			val = append(val, strconv.Itoa(bci))
		}
	}
	return strings.Join(val, "")
}

func (b *Board) getNeighbor(position *Position, d direction) *Position {
	switch d {
	case NORTH:
		return b.world.getPosition(position.X, position.Y-1)
	case SOUTH:
		return b.world.getPosition(position.X, position.Y+1)
	case WEST:
		return b.world.getPosition(position.X-1, position.Y)
	case EAST:
		return b.world.getPosition(position.X+1, position.Y)
	case NORTH_WEST:
		return b.world.getPosition(position.X-1, position.Y-1)
	case SOUTH_WEST:
		return b.world.getPosition(position.X-1, position.Y+1)
	case SOUTH_EAST:
		return b.world.getPosition(position.X+1, position.Y+1)
	case NORTH_EAST:
		return b.world.getPosition(position.X+1, position.Y-1)
	default:
		panic("unknown direction")
	}
}

func (b *Board) EmptyTileCount() int {
	memento := b.Memento()
	return strings.Count(memento, "0")
}

func (b *Board) NextBallColorsMemento() string {
	return strings.Join([]string{strconv.Itoa(b.nextBallColors[0]), strconv.Itoa(b.nextBallColors[1]), strconv.Itoa(b.nextBallColors[2])}, "")
}

func (b *Board) SetNextBallColors(memento string) {
	mementoArray := strings.Split(memento, "")
	b.nextBallColors = [3]int{}
	b.nextBallColors[0], _ = strconv.Atoi(mementoArray[0])
	b.nextBallColors[1], _ = strconv.Atoi(mementoArray[1])
	b.nextBallColors[2], _ = strconv.Atoi(mementoArray[2])
}
