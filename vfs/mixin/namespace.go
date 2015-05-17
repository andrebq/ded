package mixin

import (
	"9fans.net/go/plan9"
	"amoraes.info/ded/vfs"
	"errors"
	"path"
	"strings"
)

type (
	// Namespace should be used only by clients, since the mount information
	// and fid data is shared across all requests.
	Namespace struct {
		mounts []*Mount
		fids   map[uint32]fidPath
	}

	Mount struct {
		local         string
		remote        string
		target        vfs.RPC
		rootfid       uint32
		ctx           *vfs.Context
		fakeQid       []plan9.Qid
		qidsToDiscard int
	}

	fidPath struct {
		fullpath string
		mount    *Mount
	}
)

// Mount will expose the given plan9 fileserver, at the remote path, under our local path
//
// Example:
// Mount("/a/b", "/c/d", rpc)
//
// Will expose the file "/c/d" on rpc as if it was "/a/b" here. That means when we
// get a call for the file "/a/b/e", it will endup calling "/c/d/e" on rpc.
//
// The caller won't notice the difference since everything is handled by Namespace, for all
// intentions and purposes there is only one file called "/a/b/e".
//
// All files that live under "local" won't be available until the dir is unmounted
//t
// This works more or less like the Linux mount command
//
func (ns *Namespace) Mount(local string, remote string, rpc vfs.RPC) error {
	for _, v := range ns.mounts {
		if v.local == local {
			return errors.New("already mounted")
		}
	}
	m := &Mount{local: local, remote: remote, target: rpc, rootfid: 1, ctx: vfs.NewContext()}
	if err := m.attach(); err != nil {
		return err
	}
	ns.mounts = append(ns.mounts, m)
	return nil
}

func (ns *Namespace) Call(fc *plan9.Fcall, ctx *vfs.Context) *plan9.Fcall {
	if ns.fids == nil {
		ns.fids = make(map[uint32]fidPath)
	}
	// check if the fid already used, if that is the case,
	// just pass the call to the selected RPC
	if fp, ok := ns.fids[fc.Fid]; ok {
		// we don't use the ctx here, for nothing
		return ns.doCall(fc, fp.mount.ctx, fp.mount.target)
	}

	// check the type of the operation,
	// Tversion, Tattach don't need to be passed, since we assume that
	// the user already did those when he mounted the server
	switch fc.Type {
	case plan9.Tversion:
		ret := *fc
		ret.Type++
		ret.Msize = 8 * 1024
		ret.Version = "9P2000"
		return &ret
	case plan9.Tattach:
		ret := *fc
		ret.Type++
		ret.Qid = plan9.Qid{Path: 1}
		return &ret
	case plan9.Twalk:
		return ns.doCall(fc, ctx, nil)
	}
	ret := *fc
	return vfs.PackError(&ret, errors.New("Client MUST SEND Twalk first"))
}

func (ns *Namespace) doCall(fc *plan9.Fcall, ctx *vfs.Context, rpc vfs.RPC) *plan9.Fcall {
	switch fc.Type {
	case plan9.Twalk:
		// TODO(andre): handle multi walks
		// Problem: The current code only works if the user is capable of doing the walk
		// in a single shot. If more then one walk is required, we will give wrong results
		// but for the first prototype this should work.
		//
		// The infrastructure is already there, since we have a fid structure that holds
		// the full path (ie, the whole Wname) used to walk to that fid.
		//
		// handle with care
		name := path.Join("/", path.Join(fc.Wname...))
		if len(name) == 0 {
			name = "/"
		}
		var lastMatch *Mount
		// search all mount points and pick the one with the largest match
		// example:
		//
		// fc.Wname = "/a/b/c/d"
		// and we have mounts for "/a/b", "/a/b/c"
		// we should use the "/a/b/c" mount instead of the "/a/b"
		for _, mp := range ns.mounts {
			println("checking ", name, " against ", mp.local)
			if strings.HasPrefix(name, mp.local) {
				println("match")
				if lastMatch == nil || len(lastMatch.local) < len(mp.local) {
					lastMatch = mp
				}
			}
		}

		if lastMatch == nil {
			ret := *fc
			return vfs.PackError(&ret, errors.New("no mount point found"))
		}

		if fc.Newfid == lastMatch.rootfid {
			ret := *fc
			return vfs.PackError(&ret, errors.New("Newfid cannot be the same as the bind fid. Use another"))
		}
		// use the fid from the mount instead of the requested one
		preFid := fc.Fid
		fc.Fid = lastMatch.rootfid
		println("local to remote", lastMatch.localToRemote(name))

		fc.Wname = strings.Split(lastMatch.localToRemote(name), "/")
		println("wName", strings.Join(fc.Wname, " "))

		ret := lastMatch.target.Call(fc, lastMatch.ctx)
		fc.Fid = preFid

		if ret.Type == plan9.Rerror {
			return ret
		}

		// check if we did the full walk, if we didn't don't save the Newfid
		if len(ret.Wqid) != len(fc.Wname) {
			return ret
		}
		println("len fc.Wqid", len(ret.Wqid))
		ret.Wqid = lastMatch.remoteToLocal(ret.Wqid)
		println("len fc.Wqid (local)", len(ret.Wqid))

		// fine we completed the walk, save this to our fidmap
		ns.fids[ret.Newfid] = fidPath{
			fullpath: name,
			mount:    lastMatch,
		}

		println("saved fid path", ret.Newfid)

		return ret
	}
	if rpc == nil {
		ret := *fc
		return vfs.PackError(&ret, errors.New("didn't found a valid mount point"))
	}
	return rpc.Call(fc, ctx)
}

func (m *Mount) checkVersion() error {
	v := &plan9.Fcall{
		Type:    plan9.Tversion,
		Msize:   8 * 1024,
		Version: "9P2000",
	}
	ret := m.target.Call(v, m.ctx)
	if ret.Type == plan9.Rerror {
		return errors.New(ret.Ename)
	}
	return nil
}

func (m *Mount) attach() error {
	if err := m.checkVersion(); err != nil {
		return err
	}
	at := &plan9.Fcall{
		Type:  plan9.Tattach,
		Fid:   m.rootfid,
		Afid:  plan9.NOFID,
		Uname: "user",
		Aname: "user",
	}
	ret := m.target.Call(at, m.ctx)
	if ret.Type == plan9.Rerror {
		return errors.New(ret.Ename)
	}
	// now, let's create some fake Qids to represent "the local"
	parts := strings.Split(m.local, "/")
	m.fakeQid = make([]plan9.Qid, len(parts))
	for i, _ := range m.fakeQid {
		m.fakeQid[i].Path = uint64(i)
	}

	// now, let's count how many qids we need to discard from remote
	m.qidsToDiscard = len(strings.Split(m.remote, "/"))
	return nil
}

func (m *Mount) localToRemote(localFull string) string {
	// remove our local from localFull and concatenate with
	// the given remote
	return path.Join(m.remote, localFull[len(m.local):])
}

func (m *Mount) remoteToLocal(data []plan9.Qid) []plan9.Qid {
	ret := make([]plan9.Qid, len(m.fakeQid)+(len(data)-m.qidsToDiscard))
	copy(ret, m.fakeQid)
	copy(ret, data[m.qidsToDiscard:])
	return ret
}
