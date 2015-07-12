package namespace

import (
	"9fans.net/go/plan9"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/mixin"
	"io"
	"path"
	"time"
)

type (
	FS struct {
		mixin.FS
		ns *Namespace
	}
)

func NewFS(ns *Namespace) *FS {
	return &FS{
		ns: ns,
	}
}

func (fs *FS) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	var parent string
	if fs.ValidFid(fc, ctx) {
		if fref, ok := fs.GetFid(fc.Fid, ctx).(*FileRef); ok {
			parent = fref.AbsPath
		}
	}

	var ref FileRef
	var err error
	for _, name := range fc.Wname {
		ref, err = fs.ns.Open(path.Join("/", parent, name))
		if err != nil {
			return vfs.PackError(&ret, err)
		}
		defer ref.Close()
		dir, err := fs.fileToDir(&ref)
		if err != nil {
			return vfs.PackError(&ret, err)
		}
		ret.Wqid = append(ret.Wqid, vfs.DirToQid(dir))
		parent = ref.AbsPath
	}

	// set the newfid to the ref
	fs.SetFid(ctx, fc.Newfid, &ref)

	return &ret
}

func (fs *FS) Flush(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++
	ref := fs.GetFid(fc.Fid, ctx).(*FileRef)

	err := ref.File.Sync()
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	return &ret
}

func (fs *FS) Open(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	ref := fs.GetFid(fc.Fid, ctx).(*FileRef)

	// now, check if the ns is capable of opening the file
	actualref, err := fs.ns.Open(ref.AbsPath)
	if err != nil {
		return vfs.PackError(&ret, err)
	}

	fs.SetFid(ctx, fc.Fid, &actualref)
	return &ret
}

func (fs *FS) Read(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	buf := make([]byte, int(fc.Count))

	ref := fs.GetFid(fc.Fid, ctx).(*FileRef)
	if fc.Offset != 0 {
		_, err := ref.Seek(int64(fc.Offset), 0)
		if err != nil {
			return vfs.PackError(&ret, err)
		}
	}

	sz, err := ref.Read(buf)

	switch {
	case err == io.EOF:
		ret.Data = nil
		ret.Count = 0
		return &ret
	case err != nil:
		return vfs.PackError(&ret, err)
	}

	ret.Data = buf[0:sz]
	ret.Count = uint32(sz)
	return &ret
}

func (fs *FS) Write(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	file := fs.GetFid(fc.Fid, ctx).(*FileRef)
	if fc.Offset != 0 {
		_, err := file.Seek(int64(fc.Offset), 0)
		if err != nil {
			return vfs.PackError(&ret, err)
		}
	}

	sz, err := file.Write(fc.Data)
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	ret.Count = uint32(sz)
	return &ret
}

func (fs *FS) fileToDir(ref *FileRef) (dir plan9.Dir, err error) {
	var stat FileInfo
	stat, err = ref.Stat()
	if err != nil {
		return
	}
	if stat.IsDir {
		dir.Type |= plan9.QTDIR
	}
	dir.Atime = uint32(time.Now().Unix())
	dir.Mtime = uint32(time.Now().Unix())
	dir.Name = stat.Name
	dir.Length = uint64(stat.Size)
	dir.Uid = "none"
	dir.Gid = "none"
	dir.Muid = "none"
	return
}
