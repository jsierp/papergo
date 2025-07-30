package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const frontend = `
<html>
    <body>
        <form method="POST">
            <input type="hidden" name="action" value="left"/>
            <input type="submit" value="Left">
        </form>
        <form method="POST">
            <input type="hidden" name="action" value="right"/>
            <input type="submit" value="Right">
        </form>
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
	userId := h.authenticate(w, req)
	playerId, ok := h.playerIdMap[userId]
	if !ok {
		playerId = h.playerService.Join()
		h.playerIdMap[userId] = playerId
	}

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
	cookie, err := req.Cookie("userId")
	if errors.Is(err, http.ErrNoCookie) {
		return h.setNewUserId(w)
	}
	uid, err := uuid.Parse(cookie.Value)
	if err != nil {
		return h.setNewUserId(w)
	}
	return uid
}

func (h *HttpInputHandler) setNewUserId(w http.ResponseWriter) uuid.UUID {
	userId := uuid.New()
	http.SetCookie(w, &http.Cookie{
		Name:   "userId",
		Value:  userId.String(),
		Secure: true,
	})
	return userId
}

func (k *HttpInputHandler) Close() {
	if k.server == nil {
		return
	}
	k.server.Shutdown(context.Background())
}
