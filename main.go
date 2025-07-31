package main

func main() {
	renderer := NewRenderer()
	defer renderer.Close()

	game := NewGame(renderer)

	keyboard := NewKeyboardInputHandler(game)
	go keyboard.Listen()
	defer keyboard.Close()

	game.Run()
}
