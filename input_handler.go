package main

import (
	"golang.org/x/term"
	"log"
	"os"
)

const (
	ArrowUp    = "\033[A"
	ArrowDown  = "\033[B"
	ArrowRight = "\033[C"
	ArrowLeft  = "\033[D"
)

type KeyboardInputHandler struct {
	oldTerminalState *term.State
	playerService    PlayerService
}

func NewKeyboardInputHandler(ps PlayerService) KeyboardInputHandler {
	oldTerminalState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	return KeyboardInputHandler{
		playerService:    ps,
		oldTerminalState: oldTerminalState,
	}
}

func (k *KeyboardInputHandler) Listen() {
	buffer := make([]byte, 4)
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			log.Printf("Error reading from stdin: %v\n", err)
			break
		}
		input := string(buffer[:n])

		switch input {
		case "q", "\x03":
			log.Println("q pressed. Terminating.")
			k.playerService.Quit()
			return
		case ArrowLeft:
			k.playerService.TurnLeft(0)
		case ArrowRight:
			k.playerService.TurnRight(0)
		case "a":
			k.playerService.TurnLeft(1)
		case "d":
			k.playerService.TurnRight(1)
		default:
			//k.playerService.Join()
		}
	}

}

func (k *KeyboardInputHandler) Close() {
	term.Restore(int(os.Stdin.Fd()), k.oldTerminalState)
}
