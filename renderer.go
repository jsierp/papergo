package main

import (
	"fmt"
)

const (
	cursorTo = "\033[%d;%dH"

	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	colorReset      = "\033[0m"
	clearScreen     = "\033[2J"   // Clears the entire screen
	cursorHome      = "\033[H"    // Moves cursor to home position (1,1)
	hideCursor      = "\033[?25l" // Hides the cursor
	showCursor      = "\033[?25h" // Shows the cursor
	alternateScreen = "\033[?1049h"
	mainScreen      = "\033[?1049l"
	Solid           = "██"
	Striped         = "░░"
)

type Renderer struct {
	rows   int
	cols   int
	buffer [][]Cell
}

func NewRenderer(rows, cols int) *Renderer {
	fmt.Print(alternateScreen)
	fmt.Print(hideCursor)

	buffer := make([][]Cell, rows)
	for i := range rows {
		buffer[i] = make([]Cell, cols)
	}

	return &Renderer{rows, cols, buffer}
}

func (r *Renderer) Close() {
	fmt.Print(mainScreen)
	fmt.Print(showCursor)
}

func (r *Renderer) render(row int, col int, char string, color string) {
	fmt.Printf(cursorTo, row+1, col*2+1)
	fmt.Printf("%s%s", color, char)
}

func (r *Renderer) Refresh(g *Game) {
	frame := g.World

	for y := range r.rows {
		for x := range r.cols {
			if r.buffer[y][x] != frame[y][x] {
				r.buffer[y][x] = frame[y][x]

				if frame[y][x].TracePlayerId != 0 {
					r.render(y, x, Striped, PlayerColors[frame[y][x].TracePlayerId-1])
				} else if frame[y][x].TakenPlayerId != 0 {
					r.render(y, x, Solid, PlayerColors[frame[y][x].TakenPlayerId-1])
				} else {
					r.render(y, x, "  ", colorReset)
				}
			}
		}
	}
}
