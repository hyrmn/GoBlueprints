package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hyrmn/GoBlueprints/trace"
)

type room struct {
	//holds incoming messages
	forward chan []byte

	//channel for clients wishing to join room
	join chan *client
	//channel for clients wishing to leave room
	leave chan *client

	//clients currently in room
	clients map[*client]bool

	//our custom logger
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", string(msg))
			for client := range r.clients {
				client.send <- msg
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize}

func (r *room) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Fatalf("ServeHTTP: %s", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()

	go client.write()
	client.read()
}
