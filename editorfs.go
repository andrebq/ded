package main

import (
	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/memlistener"
	"amoraes.info/ded/vfs/mixin"
	"amoraes.info/ded/vfs/namespace"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
)

type (
	EditorFS struct {
		mixin.FS
		editor *DedEditor
		bar    *DedEditor
	}

	editorFid struct {
		name string
	}
)

func (fs *EditorFS) ExportAt(ns *namespace.Namespace, name string) error {
	var closelist []io.Closer
	ls := memlistener.New("active")
	closelist = append(closelist, ls)

	defer func(cl *[]io.Closer) {
		for _, c := range *cl {
			if c != nil {
				err := c.Close()
				if err != nil {
					log.WithFields(log.Fields{
						"Module": "EditorFS",
						"Err":    err,
					}).Errorf("Error closing resource")
				}
			}
		}
	}(&closelist)

	var err error
	fileserver := &vfs.Fileserver{fs}
	_, err = vfs.NewServer(fileserver, ls)
	if err != nil {
		println("exit 2")
		ls.Close()
		return err
	}

	nsclient, err := memlistener.Connect(ls, "nsclient")
	closelist = append(closelist, nsclient)
	if err != nil {
		println("exit 1")
		ls.Close()
		return err
	}

	conn, err := client.NewConn(nsclient)
	closelist = append(closelist, conn)
	log.Debugf("client new conn %v/%v", conn, err)
	if err != nil {
		println("exit 3")
		return err
	}

	fsys, err := conn.Attach(nil, "nouser", "nogroup")
	log.Debugf("attatch %v / %v", fsys, err)
	if err != nil {
		println("exit 4")
		return err
	}

	rootfd, err := fsys.Open("/", plan9.OREAD)
	closelist = append(closelist, rootfd)
	log.Debugf("open %v / %v", rootfd, err)
	if err != nil {
		println("exit 5")
		return err
	}

	err = ns.Mount(name, ".", rootfd)
	if err != nil {
		println("exit 6")
		return err
	}
	// nothing to close
	closelist = nil
	return nil
}

func (fs *EditorFS) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	println("editorfs walk: ", fc.String())
	ret := *fc
	ret.Type++

	if len(fc.Wname) == 0 {
		println("empty")
		fs.SetFid(ctx, fc.Newfid, &editorFid{
			name: ".",
		})
		return &ret
	}

	var efid *editorFid
	oldfd := fs.GetFid(fc.Fid, ctx)
	if oldfd != nil {
		switch oldfd := oldfd.(type) {
		case *editorFid:
			efid = oldfd
			// the only way a walk from a previous fid is valid, is if
			// the previous fid pointed to the editor itself, ie,
			// the name was "."
			if oldfd.name != "." {
				return vfs.PackError(&ret, fmt.Errorf("editors don't have subdirs"))
			}
		}
	} else {
		efid = &editorFid{}
	}

	switch fc.Wname[0] {
	case "body", "header":
		ret.Wqid = append(ret.Wqid, plan9.Qid{})
		efid.name = fc.Wname[0]
	case ".":
		ret.Wqid = append(ret.Wqid, plan9.Qid{})
		efid.name = "."
	default:
		return vfs.PackError(&ret, fmt.Errorf("file not found: %v", fc.Wname[0]))
	}
	fs.SetFid(ctx, fc.Newfid, efid)
	return &ret
}

func (fs *EditorFS) Open(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++
	fd := fs.GetFid(fc.Fid, ctx).(*editorFid)
	switch fd.name {
	case ".":
		return &ret
	default:
		return vfs.PackError(&ret, fmt.Errorf("fid %v cannot be opened", fc.Fid))
	}
}
