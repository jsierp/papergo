package main

import (
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PlayerID uint8

type Player struct {
	X         float64
	Y         float64
	Trace     bool
	MinC      int
	MinR      int
	MaxC      int
	MaxR      int
	Direction Direction
	Id        PlayerID
	Score     int
}

type Game struct {
	World         [][]Cell
	Height, Width int
	Players       map[PlayerID]*Player
	UIDToPID      map[uuid.UUID]PlayerID
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
	TakenPlayerId PlayerID
	TracePlayerId PlayerID
}

const (
	FPS           = 100
	FrameDuration = time.Second / FPS
	Speed         = 10.
	eps           = 1e-6
	FrameDelta    = 1. / FPS * Speed
)

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

func NewGame(renderer *Renderer) *Game {
	width, height := renderer.GetGameSize()
	g := &Game{
		World:    NewWorld(width, height),
		Height:   height,
		Width:    width,
		Players:  map[PlayerID]*Player{},
		UIDToPID: map[uuid.UUID]PlayerID{},
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

func (g *Game) GetPlayer(uid uuid.UUID) *Player {
	player, ok := g.Players[g.UIDToPID[uid]]
	if !ok {
		return nil
	}
	return player
}

func (g *Game) TurnLeft(u uuid.UUID) {
	player := g.GetPlayer(u)
	if player == nil {
		return
	}
	player.Direction = (player.Direction + 3) % 4
}

func (g *Game) TurnRight(u uuid.UUID) {
	player := g.GetPlayer(u)
	if player == nil {
		return
	}
	player.Direction = (player.Direction + 1) % 4
}

func (g *Game) Join(uid uuid.UUID) {
	g.playersMutex.Lock()
	defer g.playersMutex.Unlock()

	if p := g.GetPlayer(uid); p != nil {
		return
	}

	p := &Player{
		Id:        g.getMinAvailablePlayerId(),
		X:         float64(rand.Intn(g.Width)),
		Y:         float64(rand.Intn(g.Height)),
		Direction: Right,
	}

	g.claimStartingArea(p)
	g.UIDToPID[uid] = p.Id
	g.Players[p.Id] = p
}

func (g *Game) claimStartingArea(p *Player) {
	p.MinR = max(0, int(p.Y)-1)
	p.MaxR = min(int(p.Y)+1, g.Height-1)
	p.MinC = max(0, int(p.X)-1)
	p.MaxC = min(int(p.X)+1, g.Width-1)

	for row := p.MinR; row <= p.MaxR; row++ {
		for col := p.MinC; col <= p.MaxC; col++ {
			if tpID := g.World[row][col].TakenPlayerId; tpID != 0 {
				g.Players[tpID].Score--
			}
			g.World[row][col].TakenPlayerId = p.Id
			p.Score++
		}
	}
}

func (g *Game) getMinAvailablePlayerId() PlayerID {
	pIDs := make(map[PlayerID]struct{})
	for _, p := range g.Players {
		pIDs[p.Id] = struct{}{}
	}
	for i := PlayerID(1); i <= PlayerID(len(g.Players)); i++ {
		if _, found := pIDs[i]; !found {
			return i
		}
	}
	return PlayerID(len(g.Players)) + 1
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
				if tpID := g.World[p.MinR+i][p.MinC+j].TakenPlayerId; tpID != 0 {
					g.Players[tpID].Score--
				}
				g.World[p.MinR+i][p.MinC+j].TakenPlayerId = p.Id
				g.Players[p.Id].Score++
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

func (g *Game) getScoreboard() []*Player {
	scoreboard := make([]*Player, 0, len(g.Players))
	for _, p := range g.Players {
		scoreboard = append(scoreboard, p)
	}
	slices.SortFunc(scoreboard, func(a, b *Player) int {
		if a.Score == b.Score {
			return int(a.Id) - int(b.Id)
		}
		return int(b.Score) - int(a.Score)
	})
	return scoreboard
}

func (g *Game) killPlayer(pID PlayerID) {
	for uid, playerID := range g.UIDToPID {
		if pID == playerID {
			delete(g.UIDToPID, uid)
			break
		}
	}
	delete(g.Players, pID)

	for i := range g.World {
		for j := range g.World[i] {
			if g.World[i][j].TracePlayerId == pID {
				g.World[i][j].TracePlayerId = 0
			}
			if g.World[i][j].TakenPlayerId == pID {
				g.World[i][j].TakenPlayerId = 0
			}
		}
	}
}
