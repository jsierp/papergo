package main

import (
	"slices"
	"sync"
	"time"
)

type Player struct {
	X         float64
	Y         float64
	Trace     bool
	MinC      int
	MinR      int
	MaxC      int
	MaxR      int
	Direction Direction
	Id        int
	Color     string
	Dead      bool
}

type Game struct {
	World        [][]Cell
	Players      []*Player
	IsRunning    bool
	Closing      bool
	playersMutex sync.Mutex
}

type PlayerService interface {
	Join() int
	// Leave()
	ToggleIsRunning()
	Close()
	TurnLeft(int)
	TurnRight(int)
}

type CellType uint8

type Cell struct {
	PlayerId int
	Type     CellType
}

const (
	CellTypeEmpty CellType = iota
	CellTypeTrace
	CellTypeTaken
)

const Rows = 60
const Cols = 100
const FPS = 100
const FrameDuration = time.Second / FPS
const Speed = 10.
const eps = 1e-6

const FrameDelta = 1. / FPS * Speed

type Direction int

type Point struct {
	R, C int
}

const (
	Up Direction = iota
	Right
	Down
	Left
)

var PlayerDefaults []Player = []Player{
	{X: 10., Y: 10., Id: 0, Color: ColorBlue},
	{X: 40., Y: 40., Id: 1, Color: ColorRed},
}

func NewGame() *Game {
	g := &Game{
		World: NewWorld(Rows, Cols),
	}
	return g
}

func NewWorld(rows int, cols int) [][]Cell {
	world := make([][]Cell, rows)
	for i := range rows {
		world[i] = make([]Cell, cols)
	}
	return world
}

func (g *Game) Run() {
	r := NewRenderer(Rows, Cols)
	defer r.Close()

	ticker := time.NewTicker(FrameDuration)

	for !g.Closing {
		<-ticker.C
		if g.IsRunning {
			g.updatePositions()
			g.updateCells()
		}
		r.Refresh(g)
	}
}

func (g *Game) updatePositions() {
	for _, p := range g.Players {
		if !p.Dead {
			p.updatePosition()
		}
	}
}
func (p *Player) updatePosition() {
	switch p.Direction {
	case Up:
		p.Y = max(0, p.Y-FrameDelta)
	case Down:
		p.Y = min(Rows-eps, p.Y+FrameDelta)
	case Right:
		p.X = min(Cols-eps, p.X+FrameDelta)
	case Left:
		p.X = max(0, p.X-FrameDelta)
	}
}

func (g *Game) Close() {
	g.Closing = true
}

func (g *Game) TurnLeft(p int) {
	g.Players[p].Direction = (g.Players[p].Direction + 3) % 4
}

func (g *Game) TurnRight(p int) {
	g.Players[p].Direction = (g.Players[p].Direction + 1) % 4
}

func (g *Game) Join() int {
	g.playersMutex.Lock()
	pId := len(g.Players)
	g.Players = append(g.Players, &PlayerDefaults[pId])
	g.playersMutex.Unlock()

	p := g.Players[pId]
	g.World[int(p.X)][int(p.Y)].Type = CellTypeTaken
	g.World[int(p.X)][int(p.Y)].PlayerId = p.Id
	p.MinR, p.MaxR = int(p.Y), int(p.Y)
	p.MinC, p.MaxC = int(p.X), int(p.X)

	return pId
}

func (g *Game) ToggleIsRunning() {
	g.IsRunning = !g.IsRunning
}

func (g *Game) updateCells() {
	for _, p := range g.Players {
		if !p.Dead {
			g.updatePlayerCells(p)
		}
	}
}

func (g *Game) updatePlayerCells(p *Player) {
	p.updatePlayerBoundary()

	switch g.World[int(p.Y)][int(p.X)].Type {
	case CellTypeEmpty:
		g.World[int(p.Y)][int(p.X)].Type = CellTypeTrace
		g.World[int(p.Y)][int(p.X)].PlayerId = p.Id
		if !p.Trace {
			p.Trace = true
		}
	case CellTypeTaken:
		if p.Trace && g.World[int(p.Y)][int(p.X)].PlayerId == p.Id {
			g.fillTrace(p)
			p.Trace = false
		}
	case CellTypeTrace:
		if pId := g.World[int(p.Y)][int(p.X)].PlayerId; pId != p.Id {
			g.killPlayer(pId)
		}
	}
}

func (p *Player) updatePlayerBoundary() {
	p.MaxC = max(p.MaxC, int(p.X))
	p.MinC = min(p.MinC, int(p.X))
	p.MaxR = max(p.MaxR, int(p.Y))
	p.MinR = min(p.MinR, int(p.Y))
}

func (g *Game) fillTrace(p *Player) {
	takenMask := g.getTakenMask(p)
	for i := range takenMask {
		for j := range takenMask[i] {
			if takenMask[i][j] {
				g.World[p.MinR+i][p.MinC+j] = Cell{Type: CellTypeTaken, PlayerId: p.Id}
			}
		}
	}
}

func (g *Game) getTakenMask(p *Player) [][]bool {
	rows := p.MaxR - p.MinR + 1
	cols := p.MaxC - p.MinC + 1
	mask := make([][]bool, rows)
	for i := range rows {
		mask[i] = slices.Repeat([]bool{true}, cols)
	}

	var q []Point
	for i := range rows {
		considerPoint(&q, Point{i, 0}, g, p, mask)
		considerPoint(&q, Point{i, cols - 1}, g, p, mask)
	}
	for i := range cols {
		considerPoint(&q, Point{0, i}, g, p, mask)
		considerPoint(&q, Point{rows - 1, i}, g, p, mask)
	}

	for i := 0; i < len(q); i++ {
		if q[i].R > 0 {
			considerPoint(&q, Point{q[i].R - 1, q[i].C}, g, p, mask)
		}
		if q[i].C > 0 {
			considerPoint(&q, Point{q[i].R, q[i].C - 1}, g, p, mask)
		}
		if q[i].R < rows-1 {
			considerPoint(&q, Point{q[i].R + 1, q[i].C}, g, p, mask)
		}
		if q[i].C < cols-1 {
			considerPoint(&q, Point{q[i].R, q[i].C + 1}, g, p, mask)
		}
	}
	return mask
}

func considerPoint(q *[]Point, point Point, g *Game, p *Player, mask [][]bool) {
	c := g.World[p.MinR+point.R][p.MinC+point.C]
	if (c.Type == CellTypeEmpty || c.PlayerId != p.Id) && mask[point.R][point.C] {
		*q = append(*q, Point{point.R, point.C})
		mask[point.R][point.C] = false
	}
}

func (g *Game) killPlayer(pId int) {
	p := g.Players[pId]
	p.Dead = true
	for i := range g.World {
		for j := range g.World[i] {
			if g.World[i][j].PlayerId == p.Id {
				g.World[i][j].Type = CellTypeEmpty
				g.World[i][j].PlayerId = 0
			}
		}
	}
}
