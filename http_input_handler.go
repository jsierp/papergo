package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const frontend = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Full Screen Buttons</title>
    <style>
	html, body {
		height: 100%; /* Ensures html and body take full viewport height */
		margin: 0;    /* Removes default body margin */
		overflow: hidden; /* Prevents scrolling if content exceeds viewport */
	}

	.container {
		display: flex; /* Enables flexbox layout */
		height: 100%;  /* Makes the container fill the entire height of the body */
		width: 100%;   /* Makes the container fill the entire width of the body */
	}

	.button {
		width: 100%; /* Each button takes an equal share of the available space */
		height: 100%;
		border: none; /* Removes default button border */
		font-size: 2em; /* Adjust font size as needed */
		color: white; /* Text color */
		cursor: pointer; /* Indicates it's clickable */
		transition: background-color 0.3s ease; /* Smooth transition for hover effect */
	}

	.button:first-child {
		background-color: #4CAF50; /* Green for the first button */
	}

	.button:last-child {
		background-color: #008CBA; /* Blue for the second button */
	}

	.button:hover {
		opacity: 0.9; /* Slightly dim on hover */
	}
	form {
		flex: 1
	}
	</style>
	<script src="https://unpkg.com/htmx.org@1.9.12"></script>
</head>
<body>
    <div class="container">
	<form method="POST" hx-post="/" hx-swap="none">
		<input type="hidden" name="action" value="left"/>
		<input type="submit" value="Left" class="button">
	</form>
	<form method="POST" hx-post="/" hx-swap="none">
		<input type="hidden" name="action" value="right"/>
		<input type="submit" value="Right" class="button">
	</form>
    </div>
</body>
</html>`

type HttpInputHandler struct {
	playerService PlayerService
	server        *http.Server
	playerIdMap   map[uuid.UUID]int
}

func NewHttpInputHandler(ps PlayerService) *HttpInputHandler {
	handler := &HttpInputHandler{
		playerService: ps,
		playerIdMap:   map[uuid.UUID]int{},
	}

	return handler
}

func (k *HttpInputHandler) Listen(addr string) {
	k.server = &http.Server{
		Addr:    addr,
		Handler: k,
	}
	k.server.ListenAndServe()
}

func (h *HttpInputHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	playerId := h.authenticate(w, req)
	h.playerService.Join(playerId)

	action := req.FormValue("action")
	switch action {
	case "left":
		h.playerService.TurnLeft(playerId)
	case "right":
		h.playerService.TurnRight(playerId)
	}

	fmt.Fprint(w, frontend)

}

func (h *HttpInputHandler) authenticate(w http.ResponseWriter, req *http.Request) uuid.UUID {
	cookie, err := req.Cookie("playerId")
	if errors.Is(err, http.ErrNoCookie) {
		return h.setNewPlayerId(w)
	}
	uid, err := uuid.Parse(cookie.Value)
	if err != nil {
		return h.setNewPlayerId(w)
	}
	return uid
}

func (h *HttpInputHandler) setNewPlayerId(w http.ResponseWriter) uuid.UUID {
	playerId := uuid.New()
	http.SetCookie(w, &http.Cookie{
		Name:  "playerId",
		Value: playerId.String(),
	})
	return playerId
}

func (k *HttpInputHandler) Close() {
	if k.server == nil {
		return
	}
	k.server.Shutdown(context.Background())
}
