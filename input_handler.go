package main

import (
	"golang.org/x/term"
	"log"
	"os"
	"syscall"
)

const (
	ArrowUp    = "\033[A"
	ArrowDown  = "\033[B"
	ArrowRight = "\033[C"
	ArrowLeft  = "\033[D"
)

func InputHandler(onKeyPressed func(key string)) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 4)
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			log.Printf("Error reading from stdin: %v\n", err)
			break
		}
		input := string(buffer[:n])

		switch input {
		case ArrowDown, ArrowLeft, ArrowRight, ArrowUp:
			onKeyPressed(input)
		case "q":
			log.Println("q pressed. Terminating.")
			p, _ := os.FindProcess(os.Getpid())
			p.Signal(syscall.SIGINT)
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}
}
