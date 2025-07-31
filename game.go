package main

import (
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PlayerId uint8

type Player struct {
	X         float64
	Y         float64
	Trace     bool
	MinC      int
	MinR      int
	MaxC      int
	MaxR      int
	Direction Direction
	Id        PlayerId
	Color     string
}

type Game struct {
	World         [][]Cell
	Height, Width int
	Players       map[uuid.UUID]*Player
	IsRunning     bool
	Closing       bool
	playersMutex  sync.Mutex
	renderer      *Renderer
}

type PlayerService interface {
	Join(uuid.UUID)
	// Leave()
	ToggleIsRunning()
	Close()
	TurnLeft(uuid.UUID)
	TurnRight(uuid.UUID)
}

type Cell struct {
	TakenPlayerId PlayerId
	TracePlayerId PlayerId
}

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

var PlayerColors = []string{
	ColorBlue,
	ColorRed,
	ColorGreen,
	ColorYellow,
	ColorMagenta,
	ColorCyan,
	ColorWhite,
}

func NewGame(renderer *Renderer) *Game {
	width, height := renderer.GetGameSize()
	g := &Game{
		World:    NewWorld(width, height),
		Height:   height,
		Width:    width,
		Players:  map[uuid.UUID]*Player{},
		renderer: renderer,
	}
	return g
}

func NewWorld(width int, height int) [][]Cell {
	world := make([][]Cell, height)
	for i := range height {
		world[i] = make([]Cell, width)
	}
	return world
}

func (g *Game) Run() {
	ticker := time.NewTicker(FrameDuration)

	for !g.Closing {
		<-ticker.C
		if g.IsRunning {
			g.updatePositions()
			g.updateCells()
		}
		g.renderer.Refresh(g)
	}
}

func (g *Game) updatePositions() {
	for _, p := range g.Players {
		g.updatePlayerPosition(p)
	}
}
func (g *Game) updatePlayerPosition(p *Player) {
	switch p.Direction {
	case Up:
		p.Y = max(0, p.Y-FrameDelta)
	case Down:
		p.Y = min(float64(g.Height)-eps, p.Y+FrameDelta)
	case Right:
		p.X = min(float64(g.Width)-eps, p.X+FrameDelta)
	case Left:
		p.X = max(0, p.X-FrameDelta)
	}
}

func (g *Game) Close() {
	g.Closing = true
}

func (g *Game) TurnLeft(p uuid.UUID) {
	player, ok := g.Players[p]
	if !ok {
		return
	}
	player.Direction = (player.Direction + 3) % 4
}

func (g *Game) TurnRight(p uuid.UUID) {
	player, ok := g.Players[p]
	if !ok {
		return
	}
	player.Direction = (player.Direction + 1) % 4
}

func (g *Game) Join(uid uuid.UUID) {
	g.playersMutex.Lock()
	defer g.playersMutex.Unlock()

	if _, ok := g.Players[uid]; ok {
		return
	}

	p := &Player{
		Id:        g.getMinAvailablePlayerId(),
		X:         float64(rand.Intn(g.Width)),
		Y:         float64(rand.Intn(g.Height)),
		Direction: Right,
	}

	p.MinR, p.MaxR = int(p.Y), int(p.Y)
	p.MinC, p.MaxC = int(p.X), int(p.X)
	g.World[int(p.Y)][int(p.X)].TakenPlayerId = p.Id
	g.Players[uid] = p
}

func (g *Game) getMinAvailablePlayerId() PlayerId {
	pIds := make(map[PlayerId]struct{})
	for _, p := range g.Players {
		pIds[p.Id] = struct{}{}
	}
	for i := PlayerId(1); i <= PlayerId(len(g.Players)); i++ {
		if _, found := pIds[i]; !found {
			return i
		}
	}
	return PlayerId(len(g.Players)) + 1
}

func (g *Game) ToggleIsRunning() {
	g.IsRunning = !g.IsRunning
}

func (g *Game) updateCells() {
	for _, p := range g.Players {
		g.updatePlayerCells(p)
	}
}

func (g *Game) updatePlayerCells(p *Player) {
	p.updatePlayerBoundary()
	cell := &g.World[int(p.Y)][int(p.X)]

	if cell.TracePlayerId != 0 && cell.TracePlayerId != p.Id {
		g.killPlayer(cell.TracePlayerId)
	}
	if cell.TakenPlayerId == p.Id {
		if p.Trace {
			g.fillTrace(p)
			p.Trace = false
		}
	} else {
		cell.TracePlayerId = p.Id
		p.Trace = true
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
				g.World[p.MinR+i][p.MinC+j].TakenPlayerId = p.Id
			}
			if g.World[p.MinR+i][p.MinC+j].TracePlayerId == p.Id {
				g.World[p.MinR+i][p.MinC+j].TracePlayerId = 0
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
	cell := g.World[p.MinR+point.R][p.MinC+point.C]
	if cell.TakenPlayerId != p.Id && cell.TracePlayerId != p.Id && mask[point.R][point.C] {
		*q = append(*q, Point{point.R, point.C})
		mask[point.R][point.C] = false
	}
}

func (g *Game) killPlayer(pId PlayerId) {
	for uid, p := range g.Players {
		if p.Id == pId {
			delete(g.Players, uid)
			break
		}
	}

	for i := range g.World {
		for j := range g.World[i] {
			if g.World[i][j].TracePlayerId == pId {
				g.World[i][j].TracePlayerId = 0
			}
			if g.World[i][j].TakenPlayerId == pId {
				g.World[i][j].TakenPlayerId = 0
			}
		}
	}

}
