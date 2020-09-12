package handler

import (
	"net"
	"log"
)

// Handler handles
type Handler struct {
	li net.Listener
}

// New creates a new handler
func New() (*Handler) {
	li, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalln(err)
	}

	return &Handler{
		li: li,
	}
}