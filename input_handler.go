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
	CtrlC      = "\x03"
	Esc        = "\x1B"
	Space      = " "
)

type KeyboardInputHandler struct {
	oldTerminalState *term.State
	playerService    PlayerService
	arrowsPId        int
	wsadPId          int
}

func NewKeyboardInputHandler(ps PlayerService) KeyboardInputHandler {
	oldTerminalState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	return KeyboardInputHandler{
		playerService:    ps,
		oldTerminalState: oldTerminalState,
		arrowsPId:        -1,
		wsadPId:          -1,
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

		k.ensureJoined(input)

		switch input {
		case "q", CtrlC, Esc:
			log.Println("Closing the game.")
			k.playerService.Close()
			return
		case ArrowLeft:
			k.playerService.TurnLeft(k.arrowsPId)
		case ArrowRight:
			k.playerService.TurnRight(k.arrowsPId)
		case "a":
			k.playerService.TurnLeft(k.wsadPId)
		case "d":
			k.playerService.TurnRight(k.wsadPId)
		case Space:
			k.playerService.ToggleIsRunning()
		}
	}

}

func (k *KeyboardInputHandler) ensureJoined(input string) {
	switch input {
	case ArrowLeft, ArrowRight:
		if k.arrowsPId == -1 {
			k.arrowsPId = k.playerService.Join()
		}
	case "a", "d":
		if k.wsadPId == -1 {
			k.wsadPId = k.playerService.Join()
		}
	}
}

func (k *KeyboardInputHandler) Close() {
	term.Restore(int(os.Stdin.Fd()), k.oldTerminalState)
}
