package main

import (
	"log"
	"os"

	"github.com/google/uuid"
	"golang.org/x/term"
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
	arrowsPId        uuid.UUID
	wsadPId          uuid.UUID
}

func NewKeyboardInputHandler(ps PlayerService) KeyboardInputHandler {
	oldTerminalState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	return KeyboardInputHandler{
		playerService:    ps,
		oldTerminalState: oldTerminalState,
		arrowsPId:        uuid.New(),
		wsadPId:          uuid.New(),
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
		k.playerService.Join(k.arrowsPId)
	case "a", "d":
		k.playerService.Join(k.wsadPId)

	}
}

func (k *KeyboardInputHandler) Close() {
	term.Restore(int(os.Stdin.Fd()), k.oldTerminalState)
}
