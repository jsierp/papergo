package main

func main() {
	game := NewGame()

	keyboard := NewKeyboardInputHandler(game)
	go keyboard.Listen()
	defer keyboard.Close()

	httpServer := NewHttpInputHandler(game)
	defer httpServer.Close()
	go httpServer.Listen(":8080")

	game.Run()
}
