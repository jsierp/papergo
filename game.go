package main

import (
	// "log"
	"os"
	"os/signal"
	"slices"
	"syscall"
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
}

type Game struct {
	World  [][]Cell
	Player Player
}

type CellType uint8

type Cell struct {
	Player uint8
	Type   CellType
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
	Left Direction = iota
	Right
	Up
	Down
)

func NewGame() *Game {
	g := &Game{
		World:  NewWorld(Rows, Cols),
		Player: Player{X: 10., Y: 10.},
	}
	g.World[int(g.Player.X)][int(g.Player.Y)].Type = CellTypeTaken
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

	go InputHandler(g.onKeyPressed)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		r.Destroy()
		os.Exit(0)
	}()

	ticker := time.NewTicker(FrameDuration)

	for {
		<-ticker.C
		g.updatePosition()
		g.updateCells()
		r.Refresh(g.World)
	}
}

func (g *Game) onKeyPressed(key string) {
	switch key {
	case ArrowDown:
		g.Player.Direction = Down
	case ArrowUp:
		g.Player.Direction = Up
	case ArrowLeft:
		g.Player.Direction = Left
	case ArrowRight:
		g.Player.Direction = Right
	}
}

func (g *Game) updatePosition() {
	switch g.Player.Direction {
	case Up:
		g.Player.Y = max(0, g.Player.Y-FrameDelta)
	case Down:
		g.Player.Y = min(Rows-eps, g.Player.Y+FrameDelta)
	case Right:
		g.Player.X = min(Cols-eps, g.Player.X+FrameDelta)
	case Left:
		g.Player.X = max(0, g.Player.X-FrameDelta)
	}
}

func (g *Game) updateCells() {
	g.Player.updatePlayerBoundary()

	switch g.World[int(g.Player.Y)][int(g.Player.X)].Type {
	case CellTypeEmpty:
		{
			g.World[int(g.Player.Y)][int(g.Player.X)].Type = CellTypeTrace
			if !g.Player.Trace {
				g.Player.Trace = true
			}
		}
	case CellTypeTaken:
		{
			if g.Player.Trace {
				takenMask := g.getTakenMask()
				for i := range takenMask {
					for j := range takenMask[i] {
						if takenMask[i][j] {
							g.World[g.Player.MinR+i][g.Player.MinC+j] = Cell{Type: CellTypeTaken}
						}
					}
				}

				g.Player.Trace = false
			}
		}
	}
}

func (p *Player) updatePlayerBoundary() {
	p.MaxC = max(p.MaxC, int(p.X))
	p.MinC = min(p.MinC, int(p.X))
	p.MaxR = max(p.MaxR, int(p.Y))
	p.MinR = min(p.MinR, int(p.Y))
}

func (g *Game) getTakenMask() [][]bool {
	rows := g.Player.MaxR - g.Player.MinR + 1
	cols := g.Player.MaxC - g.Player.MinC + 1
	mask := make([][]bool, rows)
	for i := range rows {
		mask[i] = slices.Repeat([]bool{true}, cols)
	}

	var q []Point
	for i := range rows {
		considerPoint(&q, Point{i, 0}, g, mask)
		considerPoint(&q, Point{i, cols - 1}, g, mask)
	}
	for i := range cols {
		considerPoint(&q, Point{0, i}, g, mask)
		considerPoint(&q, Point{rows - 1, i}, g, mask)
	}

	for i := 0; i < len(q); i++ {
		if q[i].R > 0 {
			considerPoint(&q, Point{q[i].R - 1, q[i].C}, g, mask)
		}
		if q[i].C > 0 {
			considerPoint(&q, Point{q[i].R, q[i].C - 1}, g, mask)
		}
		if q[i].R < rows-1 {
			considerPoint(&q, Point{q[i].R + 1, q[i].C}, g, mask)
		}
		if q[i].C < cols-1 {
			considerPoint(&q, Point{q[i].R, q[i].C + 1}, g, mask)
		}
	}
	return mask
}

func considerPoint(q *[]Point, p Point, g *Game, mask [][]bool) {
	c := g.World[g.Player.MinR+p.R][g.Player.MinC+p.C]
	if c.Type == CellTypeEmpty && mask[p.R][p.C] {
		*q = append(*q, Point{p.R, p.C})
		mask[p.R][p.C] = false
	}
}
