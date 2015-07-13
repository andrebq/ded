package namespace

import (
	"9fans.net/go/plan9/client"
	"fmt"
)

type (
	Tree struct {
		Name   string
		Childs []*Tree
		Fid    *client.Fid
	}
)

func (t *Tree) FindChild(n string) *Tree {
	if len(n) == 0 || n == "." {
		return t
	}
	for _, c := range t.Childs {
		if c.Name == n {
			return c
		}
	}
	return nil
}

func (t *Tree) AddChild(n string) *Tree {
	nt := &Tree{
		Name: n,
	}
	t.Childs = append(t.Childs, nt)
	return nt
}

// findLongestMatch returns the deepest root in the tree that have
// a valid Fid and a tail (which could be empty) with the rest of the steps
// that should be walked using the Fid.
func (t *Tree) findLongestMatch(steps []string) (*client.Fid, []string) {
	var lastFid *client.Fid
	var lastIdx int
	child := t
	for i, s := range steps {
		println("searching child", s)
		child = child.FindChild(s)
		if child == nil {
			println("child is nil")
			println("lastIdx", lastIdx)
			println("steps", fmt.Sprintf("%v", steps[lastIdx+1:]))
			return lastFid, steps[lastIdx+1:]
		}
		if child.Fid != nil {
			lastFid = child.Fid
			lastIdx = i
			println("child has a fid.", lastFid, "lastIdx", lastIdx)
		}
	}
	println("lastIdx", lastIdx)
	return lastFid, steps[lastIdx:]
}
