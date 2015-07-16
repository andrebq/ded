package memlistener

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type (
	L struct {
		sync.Mutex
		name string

		closenotify   []chan signal
		newconns      chan net.Conn
		alreadyClosed bool
	}

	Conn struct {
		ab *pipe
		ba *pipe
	}

	pipe struct {
		local, remote simpleAddr
		r, w          *os.File
	}

	simpleAddr string

	signal struct{}
)

func newPipe() (*pipe, error) {
	p := &pipe{}
	var err error
	p.r, p.w, err = os.Pipe()
	return p, err
}

// Returns a connection to the given listener, ie,
// after this returns the connection returned is already
// accepted by somebody listening on *L.
//
// tag is used to identify the connection to the listener
func Connect(to *L, tag string) (net.Conn, error) {
	clientPipe, err := newPipe()
	if err != nil {
		return nil, err
	}
	clientPipe.local = simpleAddr(tag)
	clientPipe.remote = simpleAddr(to.name)

	serverPipe, err := newPipe()
	serverPipe.local = simpleAddr(to.name)
	serverPipe.remote = simpleAddr(tag)
	if err != nil {
		clientPipe.Close()
		return nil, err
	}
	// swap the read files to allow the server to read data written from the client
	// and vice-versa
	clientPipe.r, serverPipe.r = serverPipe.r, clientPipe.r
	select {
	// send the server pipe to the listener newconns channel
	case to.newconns <- serverPipe:
	case <-time.After(1 * time.Second):
		return nil, errors.New("timeout")
	}
	// and give the other end to the client
	return clientPipe, nil
}

func (p *pipe) Write(in []byte) (int, error) {
	return p.w.Write(in)
}

func (p *pipe) Read(out []byte) (int, error) {
	return p.r.Read(out)
}

func (p *pipe) Close() error {
	p.w.Close()
	p.r.Close()
	return nil
}

func (p *pipe) String() string {
	return fmt.Sprintf("pipe{w: %v / r: %v}", p.w, p.r)
}

// making *pipe look-like a conn
func (p *pipe) SetDeadline(t time.Time) error {
	if t != (time.Time{}) {
		return errors.New("deadline not supported")
	}
	return nil
}

func (p *pipe) SetReadDeadline(t time.Time) error {
	return p.SetDeadline(t)
}

func (p *pipe) SetWriteDeadline(t time.Time) error {
	return p.SetDeadline(t)
}

func (p *pipe) LocalAddr() net.Addr {
	return p.local
}

func (p *pipe) RemoteAddr() net.Addr {
	return p.remote
}

func New(name string) *L {
	return &L{
		name:     name,
		newconns: make(chan net.Conn),
	}
}

func (l *L) Addr() net.Addr {
	return simpleAddr(l.name)
}

func (l *L) Accept() (net.Conn, error) {
	var iAmClosed bool
	l.Lock()
	iAmClosed = l.alreadyClosed
	l.Unlock()

	if iAmClosed {
		return nil, io.EOF
	}

	ch := make(chan signal)
	l.addCloseNotify(ch)
	defer l.removeCloseNotify(ch)

	select {
	case ncon := <-l.newconns:
		return ncon, nil
	case <-ch:
		return nil, io.EOF
	}
}

func (l *L) addCloseNotify(ch chan signal) {
	l.Lock()
	defer l.Unlock()
	l.closenotify = append(l.closenotify, ch)
}

func (l *L) removeCloseNotify(ch chan signal) {
	l.Lock()
	defer l.Unlock()
	for idx, v := range l.closenotify {
		if v == ch {
			l.closenotify[idx] = nil
			if idx != len(l.closenotify)-1 {
				l.closenotify[idx] = l.closenotify[len(l.closenotify)-1]
			}
			l.closenotify = l.closenotify[0 : len(l.closenotify)-1]
			return
		}
	}
}

func (l *L) Close() error {
	l.Lock()
	l.alreadyClosed = true
	l.Unlock()

	l.Lock()
	for _, v := range l.closenotify {
		v <- signal{}
	}
	l.Unlock()
	return nil
}

func (sa simpleAddr) Network() string {
	return string(sa)
}

func (sa simpleAddr) String() string {
	return string(sa)
}
