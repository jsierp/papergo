package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func main() {
	game := NewGame(getTerminalSize())

	keyboard := NewKeyboardInputHandler(game)
	go keyboard.Listen()
	defer keyboard.Close()

	game.Run()
}

func getTerminalSize() (int, int) {
	fd := int(os.Stdin.Fd())

	// Check if the file descriptor is actually a terminal.
	if !term.IsTerminal(fd) {
		panic("Not running in a terminal.")
	}

	width, height, err := term.GetSize(fd)
	if err != nil {
		panic(fmt.Errorf("Error getting terminal size: %v", err))
	}
	return width / 2, height - 1
}
