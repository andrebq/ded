package main

import (
	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
	"amoraes.info/ded/buffer"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/memlistener"
	"amoraes.info/ded/vfs/mixin"
	"amoraes.info/ded/vfs/namespace"
	"bytes"
	"errors"
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
		name   string
		mode   uint8
		editor *DedEditor

		writer *buffer.B
		reader *bytes.Reader
	}
)

func (efid *editorFid) Close() error {
	if efid.mode == plan9.OWRITE && efid.writer != nil && efid.editor != nil {
		// flush the changes to the editor
		efid.editor.SetText(string(efid.writer.Bytes()))
	}
	return nil
}

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
	case "body", "header":
		switch fc.Mode {
		case plan9.OREAD, plan9.OWRITE:
			fd.mode = fc.Mode
		default:
			return vfs.PackError(&ret, fmt.Errorf("invalid mode"))
		}
	default:
		return vfs.PackError(&ret, fmt.Errorf("fid %v cannot be opened", fc.Fid))
	}

	switch fd.name {
	case "body":
		fd.editor = fs.editor
		if fc.Mode == plan9.OREAD {
			fd.reader = bytes.NewReader([]byte(fs.editor.Text()))
		} else {
			fd.writer = &buffer.B{}
		}
	case "header":
		fd.editor = fs.bar
		if fc.Mode == plan9.OREAD {
			fd.reader = bytes.NewReader([]byte(fs.bar.Text()))
		} else {
			fd.writer = &buffer.B{}
		}
	}
	return &ret
}

func (fs *EditorFS) Read(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++
	fd := fs.GetFid(fc.Fid, ctx).(*editorFid)
	switch fd.name {
	case ".":
		// handle the directory listing later
		return vfs.PackError(&ret, errors.New("cannot read root"))
	}

	fd.reader.Seek(int64(fc.Offset), 0)
	ret.Data = make([]byte, int(fc.Count))
	n, _ := fd.reader.Read(ret.Data)
	ret.Data = ret.Data[:n]
	ret.Count = uint32(n)
	return &ret
}

func (fs *EditorFS) Write(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++
	fd := fs.GetFid(fc.Fid, ctx).(*editorFid)

	_, err := fd.writer.Seek(int64(fc.Offset), 0)
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	n, _ := fd.writer.Write(fc.Data)
	ret.Count = uint32(n)
	return &ret
}
