package main

import (
	"9fans.net/go/plan9"
	"amoraes.info/ded/ufs"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/mixin"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	flag.Parse()

	args := flag.Args()

	ns := mixin.Namespace{}

	var toRead string

	// for each pair /local=/root
	// create a mout point
	for _, a := range args {
		if strings.Index(a, "=") > 0 {
			parts := strings.Split(a, "=")
			if len(parts) != 3 {
				fmt.Fprintf(os.Stderr, "Invalid mout for UFS. Should be <namespace local>=<remote path>=<ufs root>\n")
				return
			}
			err := ns.Mount(parts[0], parts[1], &vfs.Fileserver{&ufs.Ufs{Root: parts[2]}})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to create mount point %v\n", a)
			}
		} else {
			toRead = a
			break
		}
	}

	if len(toRead) == 0 {
		fmt.Fprintf(os.Stderr, "you MUST INFORM a file to read\n")
		os.Exit(1)
		return
	}

	// try walk
	walk := &plan9.Fcall{
		Type:   plan9.Twalk,
		Fid:    1,
		Newfid: 2,
		Wname:  strings.Split(toRead, "/"),
	}
	ctx := vfs.NewContext()
	walkRes := notError(ns.Call(walk, ctx))
	if len(walkRes.Wqid) != len(walk.Wname) {
		fmt.Printf("Walk didn't found the file. Expecting %v got %v\n", len(walk.Wname), len(walkRes.Wqid))
		os.Exit(2)
	}
	open := &plan9.Fcall{
		Type: plan9.Topen,
		Fid:  walk.Newfid,
		Mode: plan9.OREAD,
	}
	open = notError(ns.Call(open, ctx))
	read := &plan9.Fcall{
		Type:   plan9.Tread,
		Fid:    open.Fid,
		Count:  open.Iounit,
		Offset: 0,
	}
	read = notError(ns.Call(read, ctx))
	println(string(read.Data))
}

func notError(fc *plan9.Fcall) *plan9.Fcall {
	if fc.Type == plan9.Rerror {
		fmt.Fprintf(os.Stderr, "Error: %v\n", fc.Ename)
		os.Exit(1)
	}
	return fc
}
