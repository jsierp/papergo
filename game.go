package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Game struct {
	World     [][]uint8
	PlayerX   float64
	PlayerY   float64
	Direction Direction
}

const Rows = 60
const Cols = 100
const FPS = 100
const FrameDuration = time.Second / FPS
const Speed = 10.
const eps = 1e-6

const FrameDelta = 1. / FPS * Speed

type Direction int

const (
	Left Direction = iota
	Right
	Up
	Down
)

func NewGame() *Game {
	return &Game{
		World:     NewWorld(Rows, Cols),
		PlayerX:   0.,
		PlayerY:   10.,
		Direction: Right,
	}
}

func NewWorld(rows int, cols int) [][]uint8 {
	world := make([][]uint8, rows)
	for i := range rows {
		world[i] = make([]uint8, cols)
	}
	return world
}

func (g *Game) Run() {
	r := NewRenderer(g.World)

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

		r.Render(int(g.PlayerY), int(g.PlayerX), colorBlue)

		switch g.Direction {
		case Up:
			g.PlayerY = max(0, g.PlayerY-FrameDelta)
		case Down:
			g.PlayerY = min(Rows-eps, g.PlayerY+FrameDelta)
		case Right:
			g.PlayerX = min(Cols-eps, g.PlayerX+FrameDelta)
		case Left:
			g.PlayerX = max(0, g.PlayerX-FrameDelta)
		}
		r.Render(int(g.PlayerY), int(g.PlayerX), colorRed)
	}
}

func (g *Game) onKeyPressed(key string) {
	switch key {
	case ArrowDown:
		g.Direction = Down
	case ArrowUp:
		g.Direction = Up
	case ArrowLeft:
		g.Direction = Left
	case ArrowRight:
		g.Direction = Right
	}
}
