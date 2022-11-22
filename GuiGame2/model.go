package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

type wsManager struct {
	client        map[*client]bool
	register      chan *client
	unregistred   chan *client
	broadcastchan chan SocketEventStruct
}

type client struct {
	conn         *websocket.Conn
	send         chan *dataTemplate
	user         *user
	dataTemplate *dataTemplate
	mu           sync.Mutex
}

// SocketEventStruct struct of socket events
type SocketEventStruct struct {
	EventName    string      `json:"eventName"`
	EventPayload interface{} `json:"eventPayload"`
}

type user struct {
	Pseudo string `json:"pseudo"`
	ID     string `json:"id"`
}

type dataTemplate struct {
	CurrentUser user   `json:"current_user"`
	EventName   string `json:"event_name"`
	Template    string `json:"template"`
	Users       []user `json:"users"`
}
