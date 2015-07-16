package vfs

import (
	"9fans.net/go/plan9"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"runtime"
)

type (
	signal struct {
		err chan error
	}

	Server struct {
		listener net.Listener
		fs       RPC

		closeCh chan signal
		newconn chan net.Conn
		errors  chan error
	}

	Context struct {
		data map[interface{}]interface{}
	}

	RPC interface {
		Call(*plan9.Fcall, *Context) *plan9.Fcall
		ReleaseContext(*Context) error
	}

	Fileserver struct {
		ServerFS
	}

	ServerFS interface {
		Version(*plan9.Fcall, *Context) *plan9.Fcall
		Attach(*plan9.Fcall, *Context) *plan9.Fcall
		Auth(*plan9.Fcall, *Context) *plan9.Fcall
		Clunk(*plan9.Fcall, *Context) *plan9.Fcall
		Flush(*plan9.Fcall, *Context) *plan9.Fcall
		Open(*plan9.Fcall, *Context) *plan9.Fcall
		Create(*plan9.Fcall, *Context) *plan9.Fcall
		Read(*plan9.Fcall, *Context) *plan9.Fcall
		Write(*plan9.Fcall, *Context) *plan9.Fcall
		Remove(*plan9.Fcall, *Context) *plan9.Fcall
		Stat(*plan9.Fcall, *Context) *plan9.Fcall
		Wstat(*plan9.Fcall, *Context) *plan9.Fcall
		Walk(*plan9.Fcall, *Context) *plan9.Fcall

		ValidFid(*plan9.Fcall, *Context) bool
		ReleaseContext(*Context) error
	}
)

func NewTCPServer(fs RPC, bindAddr string) (*Server, error) {
	log.WithFields(log.Fields{
		"addr": bindAddr,
		"fs":   fs,
	}).Infof("Starting server")
	lst, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("Unable to start TCP server")
		return nil, err
	}
	s := &Server{
		listener: lst,
		fs:       fs,
		closeCh:  make(chan signal, 1),
		newconn:  make(chan net.Conn, 0),
		errors:   make(chan error, 1),
	}
	go s.serve()
	return s, nil
}

func NewServer(fs RPC, listener net.Listener) (*Server, error) {
	log.WithFields(log.Fields{
		"listener": listener,
		"fs":       fs,
	}).Infof("Starting server")
	s := &Server{
		listener: listener,
		fs:       fs,
		closeCh:  make(chan signal, 1),
		newconn:  make(chan net.Conn, 0),
		errors:   make(chan error, 1),
	}
	go s.serve()
	return s, nil
}

func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.errors <- err
		}
		s.newconn <- conn
	}
}

func (s *Server) serve() {
	go s.accept()
LOOP:
	for {
		select {
		case sig := <-s.closeCh:
			// TODO(andre): cleanup stuff before returning
			sig.err <- s.listener.Close()
			close(sig.err)
			s.closeCh = nil
			s.newconn = nil
			s.errors = nil
			break LOOP
		case conn := <-s.newconn:
			go s.serveConn(conn)
		case err := <-s.errors:
			if err == nil {
				continue
			}
			log.WithFields(log.Fields{
				"error":  err,
				"module": "vfs.Server",
			}).Errorf("Server Error")
		}
	}
}

func (s *Server) serveConn(conn net.Conn) {
	log.WithFields(log.Fields{
		"client": conn.RemoteAddr(),
		"module": "vfs.Server",
	}).Infof("New client connected")
	runtime.LockOSThread()
	ctx := NewContext()
	for {
		addr := conn.RemoteAddr()
		fc, err := plan9.ReadFcall(conn)
		if err != nil {
			if err != io.EOF {
				s.errors <- err
			} else {
				log.WithFields(log.Fields{
					"client": addr,
				}).Infof("Connection closed")
			}
			if err := s.fs.ReleaseContext(ctx); err != nil {
				s.errors <- err
			}

			s.errors <- conn.Close()
			break
		}
		log.WithFields(log.Fields{
			"client": addr.String(),
		}).Debugf(">> %v", fc)

		fc = s.fs.Call(fc, ctx)

		log.WithFields(log.Fields{
			"client": addr.String(),
		}).Debugf("<< %v", fc)

		if fc == nil {
			continue
		}
		plan9.WriteFcall(conn, fc)

	}
}

func (s *Server) Close() error {
	err := make(chan error)
	select {
	case s.closeCh <- signal{err: err}:
	default:
	}
	return <-err
}

func (fs *Fileserver) Call(fc *plan9.Fcall, ctx *Context) *plan9.Fcall {
	switch fc.Type {
	case plan9.Tclunk, plan9.Tflush, plan9.Twrite, plan9.Tremove,
		plan9.Tread, plan9.Tstat, plan9.Twstat, plan9.Twalk:
		if !fs.ValidFid(fc, ctx) {
			PackError(fc, ErrInvalidFid)
		}
	}
	switch fc.Type {
	case plan9.Tversion:
		fc = fs.Version(fc, ctx)
	case plan9.Tattach:
		fc = fs.Attach(fc, ctx)
	case plan9.Tclunk:
		fc = fs.Clunk(fc, ctx)
	case plan9.Tflush:
		fc = fs.Flush(fc, ctx)
	case plan9.Topen:
		fc = fs.Open(fc, ctx)
	case plan9.Tcreate:
		fc = fs.Create(fc, ctx)
	case plan9.Tread:
		fc = fs.Read(fc, ctx)
	case plan9.Twrite:
		fc = fs.Write(fc, ctx)
	case plan9.Tremove:
		fc = fs.Remove(fc, ctx)
	case plan9.Tstat:
		fc = fs.Stat(fc, ctx)
	case plan9.Twstat:
		fc = fs.Wstat(fc, ctx)
	case plan9.Twalk:
		fc = fs.Walk(fc, ctx)
	}

	return fc
}

func NewContext() *Context {
	return &Context{
		data: make(map[interface{}]interface{}),
	}
}
