package main

import (
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/namespace"
	"flag"
	log "github.com/Sirupsen/logrus"
)

type (
	sysnameHook struct {
		name string
	}
)

func (s *sysnameHook) Levels() []log.Level {
	return []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
		log.InfoLevel,
		log.DebugLevel,
	}
}

func (s *sysnameHook) Fire(e *log.Entry) error {
	e.Data["system"] = s.name
	return nil
}

var (
	addr  = flag.String("addr", ":5640", "Address to bind")
	debug = flag.Bool("debug", false, "Debug mode")
)

func main() {
	flag.Parse()
	ns := namespace.NewNamespace()
	fs := namespace.NewFS(ns)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.WithFields(log.Fields{
		"address": *addr,
	}).Infof("Starting server...")

	mdir := namespace.Memdir{}
	rdonly := namespace.NewReadOnly(`hello`, []byte(`hello`))
	wronly := namespace.NewWriteOnly(`blackhole`)
	mdir.AddFile(rdonly)
	mdir.AddFile(wronly)

	err := ns.Mount("dir", "", &mdir)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatalf("Unable to mount mdir")
	}

	srv, err := vfs.NewTCPServer(&vfs.Fileserver{fs}, *addr)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatalf("Unable to start server")
	}
	_ = srv
	select {}
}
