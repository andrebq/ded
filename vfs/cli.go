package vfs

import (
	"9fans.net/go/plan9/client"
	"log"
)

type (
	Client struct {
		*client.Conn
	}
)

const (
	defaultServer = "localhost:5640"
)

// Connect to a local 9p server running on port 5640
func ConnectLocal() (*Client, error) {
	log.Printf("Dial to %v", defaultServer)
	conn, err := client.Dial("tcp", defaultServer)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn,
	}, nil
}
