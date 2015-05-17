package main

import (
	"9fans.net/go/plan9"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/mixin"
	"errors"
	"io"
	"os"
	"time"
)

type (
	editorFs struct {
		mixin.FS
		editor    *DedEditor
		editorBar *DedEditor

		root *vfs.File
	}

	DedEditorContents struct {
		vfs.InMemory

		editor *DedEditor
	}
)

func (dec *DedEditorContents) Sync() error {
	// update the editor with the contents of the file
	dec.editor.SetTextBytes(dec.InMemory.Bytes())
	dec.editor.theme.Driver().Call(func() {
		dec.editor.Redraw()
	})
	return nil
}

func (dec *DedEditorContents) Open(mode int, _ int) (vfs.FileContents, error) {
	if mode == os.O_RDONLY {
		// if we are in read mode, copy the contents from the editor
		// to the buffer
		dec.InMemory.Seek(0, vfs.SeekStart)
		dec.InMemory.Write(dec.editor.TextBytes())
		dec.InMemory.Seek(0, vfs.SeekStart)
	}
	return dec, nil
}

func (dec *DedEditorContents) Size() (uint64, error) {
	return uint64(len(dec.editor.TextBytes())), nil
}

func (dec *DedEditorContents) Close() error {
	// force a sync on close
	dec.Sync()
	dec.InMemory.Close()
	return nil
}

func NewEditorFS(body *DedEditor, bar *DedEditor) *editorFs {
	root := vfs.NewFile("", nil)
	root.Add(vfs.NewFile("bar", &DedEditorContents{editor: bar}))
	root.Add(vfs.NewFile("body", &DedEditorContents{editor: body}))
	return &editorFs{
		editor:    body,
		editorBar: bar,
		root:      root,
	}
}

func (e *editorFs) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	if len(fc.Wname) > 1 {
		return vfs.PackError(&ret, errors.New("no dir here"))
	}

	if len(fc.Wname) == 0 {
		return &ret
	}

	file := e.root.Walk(fc.Wname[0])
	if file == nil {
		return &ret
	}
	e.SetFid(ctx, fc.Newfid, file)

	// TODO(andre): setup qid
	ret.Wqid = []plan9.Qid{plan9.Qid{}}
	return &ret
}

func (e *editorFs) Open(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	data := e.GetFid(fc.Fid, ctx).(*vfs.File)
	// TODO(andre): handle read/write/append/truncate modes
	// for now, just call open on the file
	_, err := data.Open(vfs.DirModeToOSMode(uint32(fc.Mode)), 0700)
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	ret.Qid = plan9.Qid{}
	ret.Iounit = 8 * 1024
	return &ret
}

func (e *editorFs) Write(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	file := e.GetFid(fc.Fid, ctx).(*vfs.File)
	contents, err := file.Contents()
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	_, err = contents.Seek(int64(fc.Offset), vfs.SeekStart)
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	sz, err := contents.Write(fc.Data)
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	ret.Count = uint32(sz)
	return &ret
}

func (e *editorFs) Read(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	file := e.GetFid(fc.Fid, ctx).(*vfs.File)
	contents, err := file.Contents()
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	_, err = contents.Seek(int64(fc.Offset), vfs.SeekStart)
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	ret.Data = make([]byte, int(fc.Count))

	sz, err := contents.Read(ret.Data)
	switch {
	case err == io.EOF:
		ret.Data = nil
		ret.Count = 0
	case err != nil:
		return vfs.PackError(&ret, err)
	}
	ret.Data = ret.Data[:sz]
	ret.Count = uint32(sz)
	return &ret
}

func (e *editorFs) Stat(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	dir := plan9.Dir{}
	file := e.GetFid(fc.Fid, ctx).(*vfs.File)
	sz, err := e.GetFid(fc.Fid, ctx).(*vfs.File).Size()
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	dir.Length = sz
	dir.Mode = 0644
	dir.Atime = now()
	dir.Mtime = dir.Atime
	dir.Name = file.Name
	dir.Uid = "user"
	dir.Gid = "user"
	dir.Muid = "user"
	ret.Stat, _ = dir.Bytes()
	return &ret
}

func (e *editorFs) Wstat(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	return vfs.PackError(&ret, errors.New("stat isn't supported"))
}

func now() uint32 {
	return uint32(time.Now().Unix())
}
