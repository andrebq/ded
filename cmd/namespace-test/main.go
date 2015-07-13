package main

import (
	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
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
	addr1      = flag.String("addr1", ":5640", "First address")
	addr2      = flag.String("addr2", ":5641", "Second address")
	listenAddr = flag.String("laddr", ":5642", "Address to listen for connections")
)

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)
	log.WithFields(log.Fields{
		"address":  *addr1,
		"address2": *addr2,
	}).Infof("Connecting to server...")
	fsys1, err := client.Mount("tcp", *addr1)
	if err != nil {
		log.Fatalf("Unable to connect to fsys1. %v", err)
	}
	fsys2, err := client.Mount("tcp", *addr2)
	if err != nil {
		log.Fatalf("Unable to connect to fsys2. %v", err)
	}

	fid1, err := fsys1.Open(".", plan9.OREAD)
	if err != nil {
		log.Fatalf("Unable to connect to fsys1. %v", err)
	}
	fid2, err := fsys2.Open(".", plan9.OREAD)
	if err != nil {
		log.Fatalf("Unable to connect to fsys2. %v", err)
	}

	ns := namespace.Namespace{}

	ns.Mount("fsys1", "", fid1)
	ns.Mount("fsys2", "", fid2)

	fd1, err := ns.Walk("/fsys1/main.go")
	if err != nil {
		log.Fatalf("error opening /fsys1/main.go. %v", err)
	}
	defer fd1.Close()
	fd2, err := ns.Walk("/fsys2/ufsd/main.go")
	if err != nil {
		log.Fatalf("error opening /fsys2/ufsd/main.go. %v", err)
	}
	defer fd2.Close()

	// now that we know we can connect, let's expose the namespace

	export := namespace.NewExport(&ns)
	srv, err := vfs.NewTCPServer(&vfs.Fileserver{export}, *listenAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatalf("Unable to start server")
	}
	_ = srv
	select {}
}
