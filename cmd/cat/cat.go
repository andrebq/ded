package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
)

var (
	addr  = flag.String("addr", "127.0.0.1:5640", "Address of the ded editor")
	write = flag.Bool("w", false, "Write to the file instead of reading to it. Data is consumed from stdin")
)

func main() {
	flag.Parse()
	fsys, err := client.Mount("tcp", *addr)
	if err != nil {
		panic(err)
	}
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "You must pass the filename to read from\n")
		os.Exit(1)
	}

	name := args[0]

	var fid *client.Fid
	if *write {
		ensureFileExists(fsys, name)
		fid, err := fsys.Open(name, plan9.OWRITE)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening the file: %v\n", err)
		}
		_, err = io.Copy(fid, os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing the file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fid, err := fsys.Open(name, plan9.OREAD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening the file: %v\n", err)
		}
		_, err = io.Copy(os.Stdout, fid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading the file: %v\n", err)
			os.Exit(1)
		}
	}
	err = fid.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing the file: %v\n", err)
		os.Exit(1)
	}
}

func ensureFileExists(fs *client.Fsys, name string) {
	err := fs.Access(name, plan9.AEXIST)
	if err == nil {
		// file exists, nothing to do here
		return
	}
	// maybe the file don't exist, try to create it
	fid, err := fs.Create(name, plan9.OWRITE, plan9.Perm(0644))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create file on server. %v", err)
	}
	// file created by now, release the fid and continue
	fid.Close()
}
