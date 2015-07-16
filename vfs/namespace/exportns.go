package namespace

import (
	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/mixin"
	"io"
	"path"
)

type (
	Export struct {
		mixin.FS
		ns *Namespace
	}
)

func NewExport(ns *Namespace) *Export {
	return &Export{
		ns: ns,
	}
}

func (fs *Export) isClientFid(f interface{}) bool {
	_, ok := f.(*client.Fid)
	return ok
}

func (fs *Export) Open(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	fid := fs.GetFid(fc.Fid, ctx).(*client.Fid)

	err := fid.Open(fc.Mode)
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	ret.Iounit = 8168

	return &ret
}

func (fs *Export) Create(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	fid := fs.GetFid(fc.Fid, ctx).(*client.Fid)

	err := fid.Create(fc.Name, fc.Mode, fc.Perm)
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	ret.Iounit = 8168

	return &ret
}

func (fs *Export) Write(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	fid := fs.GetFid(fc.Fid, ctx).(*client.Fid)

	sz, err := fid.WriteAt(fc.Data, int64(fc.Offset))
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	ret.Count = uint32(sz)

	return &ret
}

func (fs *Export) Read(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	fid := fs.GetFid(fc.Fid, ctx).(*client.Fid)
	buf := make([]byte, fc.Count)

	sz, err := fid.ReadAt(buf, int64(fc.Offset))
	if err != nil {
		if err == io.EOF {
			ret.Count = 0
			ret.Data = nil
			return &ret
		}
		return vfs.PackError(&ret, err)
	}
	ret.Count = uint32(sz)
	ret.Data = buf

	return &ret
}

func (fs *Export) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	oldfid := fs.GetFid(fc.Fid, ctx)
	if fs.isClientFid(oldfid) {
		println("valid fid, walk from this fid")
		// continue from the old fid found
		fid, err := oldfid.(*client.Fid).Walk(path.Join(fc.Wname...))
		if err != nil {
			return vfs.PackError(&ret, err)
		}
		oldfid.(*client.Fid).Close()
		fs.SetFid(ctx, fc.Newfid, fid)

		// should populate the Wqid here
		return &ret
	}

	println("namespace walk")
	fid, err := fs.ns.Walk(path.Join(fc.Wname...))
	if err != nil {
		println("got error")
		return vfs.PackError(&ret, err)
	}
	println("setting fid", fid)
	for _, _ = range fc.Wname {
		ret.Wqid = append(ret.Wqid, plan9.Qid{})
	}
	fs.SetFid(ctx, fc.Newfid, fid)
	return &ret
}
