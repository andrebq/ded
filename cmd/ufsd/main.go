package main

import (
	"amoraes.info/ded/ufs"
	"amoraes.info/ded/vfs"
	"flag"
	log "github.com/Sirupsen/logrus"
)

var (
	root  = flag.String("root", ".", "Default root to expose")
	addr  = flag.String("addr", ":5640", "Address to bind")
	debug = flag.Bool("debug", false, "Debug mode")
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

func init() {
	log.AddHook(&sysnameHook{
		name: "ufsd",
	})
}

func main() {
	flag.Parse()
	fs := ufs.Ufs{
		Root: *root,
	}
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.WithFields(log.Fields{
		"address": *addr,
		"root":    *root,
	}).Infof("Starting server...")

	srv, err := vfs.NewTCPServer(&vfs.Fileserver{&fs}, *addr)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatalf("Unable to start server")
	}
	_ = srv
	select {}
}
