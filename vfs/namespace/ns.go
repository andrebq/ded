package namespace

import (
	"9fans.net/go/plan9/client"
	"errors"
	"path"
	"strings"
)

type (
	Namespace struct {
		mounts Tree
	}
)

// Mount changes this namespace to expose fid (and it's tree) under parent/name.
func (ns *Namespace) Mount(name string, parent string, fid *client.Fid) error {
	root := &ns.mounts
	for _, p := range strings.Split(parent, "/") {
		child := root.FindChild(p)
		if child == nil {
			child = root.AddChild(p)
		}
		root = child
	}
	if root.FindChild(name) != nil {
		return errors.New("name is duplicated")
	}
	root = root.AddChild(name)
	root.Fid = fid
	return nil
}

// Walk scans the mount tree for the path p and perform a walk on the correct fid.
//
// When walk reaches a node without a valid child, then it will start to perform
// walk operations on the last valid fid.
func (ns *Namespace) Walk(p string) (*client.Fid, error) {
	var searchFid *client.Fid
	root := &ns.mounts
	p = path.Join("/", p)
	parts := strings.Split(p, "/")
	fid, parts := root.findLongestMatch(parts)

	if fid == nil {
		return nil, errors.New("path not found")
	}
	if len(parts) > 0 {
		// need to do fid walking
		var err error
		// clone the fid from the mount point
		searchFid, err = fid.Walk(".")
		if err != nil {
			return nil, err
		}
		defer searchFid.Close()
		for i, s := range parts {
			searchFid, err = searchFid.Walk(s)
			if err != nil {
				return nil, err
			}
			// if this isn't the last fid to be scanned
			// close it to release the resources.
			//
			// The last one shouldn't be closed, since it will
			// be used by the caller of Walk
			if i != len(parts)-1 {
				defer searchFid.Close()
			}
		}
	} else {
		// no more walking, just clone the fid
		var err error
		searchFid, err = fid.Walk(".")
		if err != nil {
			return nil, err
		}
	}
	return searchFid, nil
}
