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
	World [][]uint8
}

func NewRenderer(world [][]uint8) *Renderer {
	fmt.Print(alternateScreen)
	fmt.Print(hideCursor)

	return &Renderer{World: world}
}

func (r *Renderer) Destroy() {
	fmt.Print(mainScreen)
	fmt.Print(showCursor)
}

func (r *Renderer) Render(row int, col int, char string, color string) {
	fmt.Printf(cursorTo, row+1, col*2+1)
	fmt.Printf("%s%s", color, char)
}
