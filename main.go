package main

func main() {
	game := NewGame()

	keyboard := NewKeyboardInputHandler(game)
	go keyboard.Listen()
	defer keyboard.Close()

	game.Run()
}
