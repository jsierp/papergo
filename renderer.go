package main

import "fmt"

const (
	cursorTo        = "\033[%d;%dH"
	colorBlue       = "\033[34m"
	colorRed        = "\033[31m"
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
	buffer [][]uint8
}

func NewRenderer(rows, cols int) *Renderer {
	fmt.Print(alternateScreen)
	fmt.Print(hideCursor)

	buffer := make([][]uint8, rows)
	for i := range rows {
		buffer[i] = make([]uint8, cols)
	}

	return &Renderer{rows, cols, buffer}
}

func (r *Renderer) Destroy() {
	fmt.Print(mainScreen)
	fmt.Print(showCursor)
}

func (r *Renderer) render(row int, col int, char string, color string) {
	fmt.Printf(cursorTo, row+1, col*2+1)
	fmt.Printf("%s%s", color, char)
}

func (r *Renderer) Refresh(frame [][]uint8) {
	for y := range r.rows {
		for x := range r.cols {
			if r.buffer[y][x] != frame[y][x] {
				r.buffer[y][x] = frame[y][x]
				r.render(y, x, Solid, colorBlue)
			}

		}

	}

}
