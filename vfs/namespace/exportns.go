package namespace

import (
	"9fans.net/go/plan9"
	"9fans.net/go/plan9/client"
	"amoraes.info/ded/vfs"
	"amoraes.info/ded/vfs/mixin"
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

func (fs *Export) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++

	oldfid := fs.GetFid(fc.Fid, ctx)
	if oldfid != nil {
		// continue from the old fid found
		fid, err := oldfid.(*client.Fid).Walk(path.Join(fc.Wname...))
		if err != nil {
			return vfs.PackError(&ret, err)
		}
		oldfid.(*client.Fid).Close()
		fs.SetFid(ctx, fc.Fid, fid)

		// should populate the Wqid here
		return &ret
	}

	fid, err := fs.ns.Walk(path.Join(fc.Wname...))
	if err != nil {
		return vfs.PackError(&ret, err)
	}
	fs.SetFid(ctx, fc.Fid, fid)
	return &ret
}
