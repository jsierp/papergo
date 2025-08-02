package main

func main() {
	renderer := NewRenderer()
	defer renderer.Close()

	game := NewGame(renderer)

	keyboard := NewKeyboardInputHandler(game)
	go keyboard.Listen()
	defer keyboard.Close()

	httpServer := NewHttpInputHandler(game)
	defer httpServer.Close()
	go httpServer.Listen(":8080")

	game.Run()
}
