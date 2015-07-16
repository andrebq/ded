package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"

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
		fmt.Fprintf(os.Stderr, "error mounting: %v", err)
		os.Exit(1)
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
		var err error
		fid, err = fsys.Open(name, plan9.OWRITE)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening the file: %v\n", err)
			os.Exit(1)
		}
		_, err = io.Copy(fid, os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing the file: %v\n", err)
			os.Exit(1)
		}
	} else {
		var err error
		fid, err = fsys.Open(name, plan9.OREAD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening the file: %v\n", err)
			os.Exit(1)
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

func ensureFileExists(fsys *client.Fsys, name string) {
	fid, err := fsys.Open(name, plan9.OWRITE)
	if err == nil {
		fid.Close()
		// file exists, nothing to do here
		return
	}
	pdir, fname := path.Split(name)
	// maybe the file don't exist, try to create it
	// at the parent location
	fid, err = fsys.Open(pdir, plan9.OREAD)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to navigate to parent directory. %v\n", err)
		os.Exit(1)
	}
	// no need to hang with parent directory
	defer fid.Close()

	err = fid.Create(fname, plan9.OWRITE, plan9.Perm(0644))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create file on server. %v\n", err)
		os.Exit(1)
	}
	// file created by now, release the fid and continue
	fid.Close()
}
