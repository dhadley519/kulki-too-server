package board

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

type direction int

type Board interface {
	Get(position *Position) int
	randomNext3Colors()
	addNext3Balls() []AddBall
	Depth() int
	Width() int
	Set(position *Position, color int)
	aStar(start *Position, goal *Position) ([]*Position, error)
}

type boardInternal struct {
	data           [][]int
	numberOfColors int
	nextBallColors [3]int
}

type MovePath struct {
	from *Position
	to   *Position
}

type Position struct {
	x int
	y int
}

type Color struct {
	color int
}

type CommandType string

type Command interface {
	getType() CommandType
}

type AddBall struct {
	Position
	Color
}

type RemoveBall struct {
	Position
}

const (
	ADD    CommandType = "ADD"
	REMOVE CommandType = "REMOVE"
)

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

func getMovementDirections() []direction {
	return []direction{NORTH, SOUTH, WEST, EAST}
}

func getDirections() []direction {
	return []direction{NORTH, SOUTH, WEST, EAST, NORTH_WEST, SOUTH_WEST, SOUTH_EAST, NORTH_EAST}
}

func (b boardInternal) Depth() int {
	return len(b.data)
}

func (b boardInternal) Width() int {
	return len(b.data[0])
}

func (b boardInternal) Get(position *Position) int {
	if position == nil {
		panic("requested position of nil")
	}
	if position.x <= b.Width() && position.y <= b.Depth() {
		color := b.data[position.y][position.x]
		return color
	}
	panic("inconsistent board size or request")
}

func (b boardInternal) randomNext3Colors() {
	for i := 0; i < 3; i++ {
		b.nextBallColors[i] = rand.Intn(b.numberOfColors-1) + 1
	}
}

func (b boardInternal) Set(position *Position, color int) {
	if color == 0 {
		//attempt to release tile
		b.data[position.y][position.x] = 0
		return
	}
	if b.Get(position) == 0 {
		//empty tile gets color
		b.data[position.y][position.x] = color
		return
	}
	log.Println("attempt to set color of non empty tile")
}

func (b boardInternal) addNext3Balls() []AddBall {
	var commands []AddBall
	for i := 0; i < 3; {
		y := rand.Intn(b.Depth())
		x := rand.Intn(b.Width())
		p := &Position{x, y}
		ball := b.Get(p)
		if ball == 0 {
			color := b.nextBallColors[i]
			b.Set(p, color)
			commands = append(commands, AddBall{Position{x: x, y: y}, Color{color: color}})
			i++
		}
	}
	b.randomNext3Colors()
	return commands
}

func (b boardInternal) onStartMessageReceipt() []AddBall {
	b.randomNext3Colors()
	return b.addNext3Balls()
}

func (b boardInternal) onMoveMessageReceipt(move *MovePath) []Command {
	success := b.move(move)
	var changes []Command
	if success {
		color := b.Get(move.to)

		changes = append(changes, AddBall{*(*move).to, Color{color}})
		changes = append(changes, RemoveBall{*(*move).from})
		removals := b.getMatchRemovalsAndUpdateScore(move.from)
		if len(removals) > 0 {
			for _, el := range removals {
				changes = append(changes, el)
			}
		} else {
			adds := b.addNext3Balls()
			for _, a := range adds {
				changes = append(changes, a)
				removals := b.getMatchRemovalsAndUpdateScore(&a.Position)
				for _, r := range removals {
					changes = append(changes, r)
				}
			}
		}
	}
	return changes
}

func (b boardInternal) onPathFindReceipt(path *MovePath) {

}

func (b boardInternal) move(move *MovePath) bool {
	color := b.Get(move.from)
	if color != 0 {
		b.release(move.from)
		b.Set(move.to, color)
		return true
	}
	log.Println("failed to move")
	return false
}

func (b boardInternal) release(position *Position) bool {
	color := b.Get(position)
	if color != 0 {
		b.Set(position, 0)
		return true
	}
	log.Println("failed to release")
	return false
}

func (b boardInternal) getMatchRemovalsAndUpdateScore(position *Position) []RemoveBall {
	return []RemoveBall{}
}

func NewBoard(width int, depth int, numberOfColors int) Board {
	b := boardInternal{data: make([][]int, depth),
		numberOfColors: numberOfColors,
		nextBallColors: [3]int{0, 0, 0},
	}
	for i := 0; i < depth; i++ {
		b.data[i] = make([]int, width)
	}
	return b
}

func (c AddBall) getType() CommandType {
	return ADD
}

func (c RemoveBall) getType() CommandType {
	return REMOVE
}

func abs(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}

/** number of hops between start and goal with a pinch of random fractional part to help the algorithm randomly select one way over another equivalent way */
func manhattan(start *Position, goal *Position) float32 {
	return float32(abs(start.x-goal.x)+abs(start.y-goal.y)) + rand.Float32()
}

func stringValue(p *Position) string {
	return fmt.Sprint(p.x, p.y)
}

func fromString(s string) *Position {
	fields := strings.Fields(s)
	x, _ := strconv.Atoi(fields[0])
	y, _ := strconv.Atoi(fields[1])
	return &Position{x, y}
}

func (b boardInternal) fillFScore(fScore map[string]float32, goal *Position) map[string]float32 {
	for i := 0; i < b.Depth(); i++ {
		for j := 0; j < b.Width(); j++ {
			p := &Position{j, i}
			fScore[stringValue(p)] = manhattan(p, goal)
		}
	}
	return fScore
}

func (b boardInternal) aStar(start *Position, goal *Position) ([]*Position, error) {
	openSet := []*Position{start}
	cameFrom := make(map[*Position]*Position)
	gScore := make(map[string]float32)
	gScore[stringValue(start)] = 0
	fScore := b.fillFScore(make(map[string]float32), goal)
	counter := 0
	for len(openSet) > 0 {
		counter++
		current := lowest(openSet, fScore)
		if current.equal(goal) {
			path := reconstructPath(cameFrom, current)
			for _, position := range path {
				println(fmt.Sprint(position.x, position.y))
			}
			println("iterations: ", counter)
			return path, nil
		}
		neighbors := b.neighborsWithoutBalls(current, cameFrom)
		openSet = remove(openSet, current)
		for _, neighbor := range neighbors {
			tentative_gScore := value(gScore, current) + 1
			if tentative_gScore < value(gScore, neighbor) {
				cameFrom[neighbor] = current
				gScore[stringValue(neighbor)] = tentative_gScore
				fScore[stringValue(neighbor)] = tentative_gScore + manhattan(start, neighbor)
				if !contains(openSet, neighbor) {
					openSet = append(openSet, neighbor)
				}
			}
		}
	}
	return nil, errors.New("path not found")
}

func value(m map[string]float32, p *Position) float32 {
	val, b := m[stringValue(p)]
	if !b {
		return math.MaxFloat32
	}
	return val
}

func contains(set []*Position, neighbor *Position) bool {
	for _, position := range set {
		if position.equal(neighbor) {
			return true
		}
	}
	return false
}

func (b boardInternal) neighborsWithoutBalls(point *Position, cameFrom map[*Position]*Position) []*Position {
	var neighbors []*Position
	directions := getMovementDirections()

	for _, direction := range directions {
		neighbor := b.getNeighbor(point, direction)
		backtrack := append(keys(cameFrom), values(cameFrom)...)
		if neighbor != nil && !b.isSet(neighbor) && !contains(backtrack, neighbor) {
			neighbors = append(neighbors, neighbor)
		}
	}
	return neighbors
}

func keys(m map[*Position]*Position) []*Position {
	var keys []*Position
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func values(m map[*Position]*Position) []*Position {
	var values []*Position
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func (b boardInternal) getNeighbor(position *Position, d direction) *Position {
	var neighbor Position
	switch d {
	case NORTH:
		neighbor = Position{x: position.x, y: position.y - 1}
		break
	case SOUTH:
		neighbor = Position{x: position.x, y: position.y + 1}
		break
	case WEST:
		neighbor = Position{x: position.x - 1, y: position.y}
		break
	case EAST:
		neighbor = Position{x: position.x + 1, y: position.y}
		break
	case NORTH_WEST:
		neighbor = Position{x: position.x - 1, y: position.y - 1}
		break
	case SOUTH_WEST:
		neighbor = Position{x: position.x - 1, y: position.y + 1}
		break
	case SOUTH_EAST:
		neighbor = Position{x: position.x + 1, y: position.y + 1}
		break
	case NORTH_EAST:
		neighbor = Position{x: position.x + 1, y: position.y - 1}
		break
	}
	if b.isPositionOnTheBoard(&neighbor) {
		return &neighbor
	}
	return nil
}

func (b boardInternal) isPositionOnTheBoard(p *Position) bool {
	return b.Width() > p.x && b.Depth() > p.y && p.x >= 0 && p.y >= 0
}

func (b boardInternal) isSet(position *Position) bool {
	if position == nil {
		panic("is set called with nil")
	}
	return b.Get(position) > 0
}

func remove(openSet []*Position, current *Position) []*Position {
	var result []*Position
	for _, p := range openSet {
		if !p.equal(current) {
			result = append(result, p)
		}
	}
	return result
}

func reconstructPath(cameFrom map[*Position]*Position, current *Position) []*Position {
	var path = []*Position{current}
	for contains(keys(cameFrom), current) {
		current = cameFrom[current]
		path = append([]*Position{current}, path...)
	}
	return path
}

func (pointA Position) equal(pointB *Position) bool {
	return pointA.x == pointB.x && pointA.y == pointB.y
}

/* return the node in openSet having the lowest fScore[] value */
func lowest(openSet []*Position, fScore map[string]float32) *Position {

	if len(openSet) == 1 {
		return openSet[0]
	}

	current := openSet[0]
	estimate := value(fScore, current)

	for i, point := range openSet {
		if i == 0 {
			continue
		}
		score := value(fScore, point)
		if score < estimate {
			current = point
			estimate = score
		}
	}
	return current
}
