package mixin

import (
	"9fans.net/go/plan9"
	"amoraes.info/ded/vfs"
	log "github.com/Sirupsen/logrus"
	"io"
)

type (
	FS struct{}

	keys uint

	fidMap map[uint32]interface{}
)

const (
	fids = keys(iota)
)

func (fs *FS) GetFid(fid uint32, ctx *vfs.Context) interface{} {
	return (ctx.MustGet(fids).(fidMap))[fid]
}

func (fs *FS) SetFid(ctx *vfs.Context, fid uint32, val interface{}) {
	ctx.MustGet(fids).(fidMap)[fid] = val
}

func (fs *FS) ValidFid(fc *plan9.Fcall, ctx *vfs.Context) bool {
	return fs.GetFid(fc.Fid, ctx) != nil
}

func (fs *FS) ReleaseFid(fid uint32, ctx *vfs.Context) error {
	log.WithFields(log.Fields{
		"Fid":    fid,
		"Module": "mixin.FS",
	}).Debugf("Releasing fid")
	defer func(ctx *vfs.Context, fid uint32) { delete((ctx.MustGet(fids).(fidMap)), fid) }(ctx, fid)
	fd := fs.GetFid(fid, ctx)
	if fd != nil {
		switch fd := fd.(type) {
		case io.Closer:
			log.WithFields(log.Fields{
				"Fid":    fid,
				"Module": "mixin.FS",
			}).Debugf("Fid is a Closer")
			err := fd.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *FS) Version(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ctx.Put(fids, make(fidMap))
	ret := *fc
	ret.Type++
	ret.Msize = 1024 * 8
	ret.Version = "9P2000"
	return &ret
}

func (fs *FS) Attach(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++
	fs.SetFid(ctx, fc.Fid, struct{}{})
	return &ret
}

func (fs *FS) Auth(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type++
	return &ret
}

func (fs *FS) Clunk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	println("cluking fid", fc.Fid)
	ret := *fc
	ret.Type++
	if err := fs.ReleaseFid(fc.Fid, ctx); err != nil {
		println("error on release", err.Error())
		return vfs.PackError(&ret, err)
	}
	return &ret
}

func (fs *FS) Flush(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Open(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Create(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Read(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Write(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Remove(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Stat(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Wstat(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) Walk(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	ret := *fc
	ret.Type = plan9.Rerror
	ret.Ename = "filesystem missing implementation"
	return &ret
}

func (fs *FS) ReleaseContext(ctx *vfs.Context) error {
	log.WithFields(log.Fields{
		"Module": "mixin.FS",
	}).Debugf("Releasing context data")
	if _, ok := ctx.Get(fids); !ok {
		log.WithFields(log.Fields{
			"Module": "mixin.FS",
		}).Debugf("No fid to release")
		return nil
	}
	fids := ctx.MustGet(fids).(fidMap)
	var firstErr error
	for k, _ := range fids {
		err := fs.ReleaseFid(k, ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"Fid":    k,
				"Module": "mixin.FS",
				"Err":    err,
			}).Errorf("ReleaseFid error")
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
