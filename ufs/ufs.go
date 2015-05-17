package ufs

import (
	"9fans.net/go/plan9"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/mixin"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type (
	Ufs struct {
		mixin.FS
		Root string
	}

	ufsFid struct {
		fullpath string
		file     *os.File
	}
)

func (fd *ufsFid) Close() error {
	if fd.file != nil {
		return fd.file.Close()
	}
	return nil
}

func (ufs *Ufs) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	parent := ufs.Root
	for _, name := range fc.Wname {
		file, err := os.Stat(filepath.Join(parent, name))
		if err != nil {
			return vfs.PackError(&ret, err)
		}
		dir := FileInfoToDir(file)
		ret.Wqid = append(ret.Wqid, DirToQid(dir))
		parent = filepath.Join(parent, name)
	}

	ufs.SetFid(ctx, fc.Newfid, &ufsFid{
		fullpath: parent,
	})

	return &ret
}

func (ufs *Ufs) Open(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	data := ufs.GetFid(fc.Fid, ctx).(*ufsFid)
	openmode := DirModeToOSMode(uint32(fc.Mode))
	file, err := os.OpenFile(data.fullpath, openmode, 0666)
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	data.file = file
	ret.Iounit = 8 * 1024

	return &ret
}

func (ufs *Ufs) Create(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	data := ufs.GetFid(fc.Fid, ctx).(*ufsFid)
	openmode := DirModeToOSMode(uint32(fc.Mode))

	perm := fc.Perm

	unixPerm := vfs.Plan9PermToUnix(uint32(fc.Perm))
	if (perm & plan9.DMDIR) == plan9.DMDIR {
		// we are creating a directory
		// the only valid permission will be 0755
		if unixPerm != 0755 {
			return vfs.PackError(&ret, errors.New("Only 0644 (file) and 0755 (dir) permissions are allowed"))
		}
	} else {
		if unixPerm != 0644 {
			return vfs.PackError(&ret, errors.New("Only 0644 (file) and 0755 (dir) permissions are allowed"))
		}
	}

	file, err := os.OpenFile(data.fullpath, openmode|os.O_CREATE, unixPerm)
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	data.file = file
	ret.Iounit = 8 * 1024

	return &ret
}

func (ufs *Ufs) Read(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	buf := make([]byte, int(fc.Count))

	file := ufs.GetFid(fc.Fid, ctx).(*ufsFid).file
	if fc.Offset != 0 {
		_, err := file.Seek(int64(fc.Offset), 0)
		if err != nil {
			return vfs.PackError(&ret, err)
		}
	}

	sz, err := file.Read(buf)

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

func (ufs *Ufs) Write(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	file := ufs.GetFid(fc.Fid, ctx).(*ufsFid).file
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

func (ufs *Ufs) Flush(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	file := ufs.GetFid(fc.Fid, ctx).(*ufsFid).file

	err := file.Sync()
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	return &ret
}

func FileInfoToDir(stat os.FileInfo) (dir plan9.Dir) {
	return vfs.FileInfoToDir(stat)
}

func DirToQid(dir plan9.Dir) plan9.Qid {
	return vfs.DirToQid(dir)
}

func DirModeToOSMode(dm uint32) (osmode int) {
	return vfs.DirModeToOSMode(dm)
}
