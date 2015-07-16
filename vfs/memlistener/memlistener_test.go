package memlistener

import (
	"bytes"
	"testing"
)

func TestListener(t *testing.T) {
	l := New("server")
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Fatalf("Error getting a new connection %v", err)
		}
		if conn.RemoteAddr().Network() != "client" {
			t.Fatalf("invalid remoteaddr name for server conn")
		}
		if conn.LocalAddr().Network() != "server" {
			t.Fatalf("invalid localaddr name for server conn")
		}
		buf := make([]byte, 4)
		_, err = conn.Read(buf)
		if err != nil {
			t.Fatalf("unable to read from client %v", err)
		}
		if !bytes.Equal(buf, []byte("helo")) {
			t.Fatalf("Invalid message from connection. %v", string(buf))
		}
		_, err = conn.Write([]byte("oleh"))
		if err != nil {
			t.Fatalf("Error sending data to client %v", err)
		}

		conn.Close()
	}()

	cli, err := Connect(l, "client")
	if err != nil {
		t.Fatalf("unable to connect to server %v", err)
	}

	if cli.RemoteAddr().Network() != "server" {
		t.Fatalf("invalid remoteaddr for client conn")
	}

	if cli.LocalAddr().Network() != "client" {
		t.Fatalf("invalid localaddr for client conn")
	}

	_, err = cli.Write([]byte("helo"))
	if err != nil {
		t.Fatalf("unable to write to server %v", err)
	}

	buf := make([]byte, 4)
	_, err = cli.Read(buf)
	if err != nil {
		t.Fatalf("unable to read from server %v", err)
	}
	if !bytes.Equal(buf, []byte("oleh")) {
		t.Fatalf("invalid read from server")
	}
}
