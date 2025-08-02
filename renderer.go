package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

const (
	cursorTo = "\033[%d;%dH"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	colorReset      = "\033[0m"
	clearScreen     = "\033[2J"   // Clears the entire screen
	cursorHome      = "\033[H"    // Moves cursor to home position (1,1)
	hideCursor      = "\033[?25l" // Hides the cursor
	showCursor      = "\033[?25h" // Shows the cursor
	alternateScreen = "\033[?1049h"
	mainScreen      = "\033[?1049l"
	Solid           = '▓'
	Striped         = '░'
	Head            = '█'
)

type colorID uint8

var PlayerColors = [...]string{
	Black,
	Red,
	Green,
	Yellow,
	Blue,
	Magenta,
	Cyan,
	White,
	BrightRed,
	BrightGreen,
	BrightYellow,
	BrightBlue,
	BrightMagenta,
	BrightCyan,
	BrightWhite,
	BrightBlack,
}

var PlayerNames = [...]string{
	"Blakko", // Black
	"Reddy",  // Red
	"Greeno", // Green
	"Yelloz", // Yellow
	"Bluppo", // Blue
	"Maggo",  // Magenta
	"Cynic",  // Cyan
	"Whizzo", // White
	"Zedred", // BrightRed
	"Grinko", // BrightGreen
	"Yellix", // BrightYellow
	"Bluzo",  // BrightBlue
	"Magito", // BrightMagenta
	"Cyzzy",  // BrightCyan
	"Whitzo", // BrightWhite
	"Shadok", // BrightBlack
}

const scoreboardWidth = 15

type character struct {
	char    rune
	colorID colorID
}

type Renderer struct {
	gameCols       int
	gameRows       int
	buffer         [][]character
	terminalWidth  int
	terminalHeight int
	scorebaord     bool
}

func NewRenderer() *Renderer {
	fmt.Print(alternateScreen)
	fmt.Print(hideCursor)

	width, height := getTerminalSize()

	gameCols, gameRows := width/2, height

	if width > height && width > 100 {
		gameCols -= scoreboardWidth
	} else {
		log.Println("Terminal too small for scoreboard, hiding it.")
	}

	r := Renderer{
		gameCols:       gameCols,
		gameRows:       gameRows,
		terminalWidth:  width,
		terminalHeight: height,
	}
	r.buffer = r.newBuffer()
	return &r
}

func (r *Renderer) Close() {
	fmt.Print(mainScreen)
	fmt.Print(showCursor)
}

func (r *Renderer) render(row int, col int, char string, color string) {
	fmt.Printf(cursorTo, row, col)
	fmt.Printf("%s%s", color, char)
}

func (r *Renderer) newBuffer() [][]character {
	buffer := make([][]character, r.terminalHeight)
	for i := range r.terminalHeight {
		buffer[i] = make([]character, r.terminalWidth)
	}
	return buffer
}

func (r *Renderer) Refresh(g *Game) {
	buffer := r.newBuffer()
	r.bufferGame(g, buffer)
	r.bufferHeads(g, buffer)
	r.bufferScoreboard(g, buffer)

	for y := 0; y < r.terminalHeight; y++ {
		for x := 0; x < r.terminalWidth; x++ {
			if buffer[y][x] != r.buffer[y][x] {
				if buffer[y][x].char == 0 {
					r.render(y, x, " ", colorReset)
				} else {
					r.render(y, x, string(buffer[y][x].char), PlayerColors[buffer[y][x].colorID])
				}
			}
		}
	}
	r.buffer = buffer
}

func (r *Renderer) bufferScoreboard(g *Game, b [][]character) {
	scoreboard := g.getScoreboard()

	for j, char := range "SCOREBOARD" {
		if j < scoreboardWidth-1 {
			b[1][r.gameCols*2+2+j] = character{char: char, colorID: 7}
		}
	}

	for i, p := range scoreboard {
		log.Println(p.Id)
		name := PlayerNames[p.Id]
		scoreText := fmt.Sprintf("%s: %d", name, 100)
		for j, char := range scoreText {
			if j < scoreboardWidth-1 {
				b[i*2+3][r.gameCols*2+2+j] = character{char: char, colorID: colorID(p.Id)}
			}
		}
	}
}

func (r *Renderer) bufferGame(g *Game, b [][]character) {
	for y := 0; y < r.gameRows; y++ {
		for x := 0; x < r.gameCols; x++ {
			var c character
			if g.World[y][x].TracePlayerId != 0 {
				c = character{
					char:    Striped,
					colorID: colorID(g.World[y][x].TracePlayerId),
				}
			} else if g.World[y][x].TakenPlayerId != 0 {
				c = character{
					char:    Solid,
					colorID: colorID(g.World[y][x].TakenPlayerId),
				}
			} else {
				c = character{
					char:    ' ',
					colorID: 0,
				}
			}
			b[y][x*2] = c
			b[y][x*2+1] = c
		}
	}
}

func (r *Renderer) bufferHeads(g *Game, b [][]character) {
	for _, p := range g.Players {
		c := character{
			char:    Head,
			colorID: colorID(p.Id),
		}
		b[int(p.Y)][int(p.X)*2] = c
		b[int(p.Y)][int(p.X)*2+1] = c
	}
}

func getTerminalSize() (width, height int) {
	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		panic("Not running in a terminal.")
	}

	width, height, err := term.GetSize(fd)
	if err != nil {
		panic(fmt.Errorf("Error getting terminal size: %v", err))
	}
	return
}

func (r *Renderer) GetGameSize() (width, height int) {
	return r.gameCols, r.gameRows
}
