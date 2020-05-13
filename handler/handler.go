package handler

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type handler struct {
	cache *Cache
}

func NewHandler(c *Cache) handler {
	return handler{c}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "websocket upgrade error", http.StatusBadRequest)
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		messageType, b, err := conn.ReadMessage()
		if err != nil || messageType != websocket.TextMessage {
			return
		}
		key := string(b)
		if h.cache.Validate(key) {
			h.cache.Incr(key)
		}
	}
}
