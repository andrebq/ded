package namespace

import (
	"9fans.net/go/plan9/client"
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
	if len(steps) == 1 {
		return t.Fid, nil
	}
	first := steps[0]
	tail := steps[1:]

	var validChild *Tree
	if len(first) == 0 || first == "." || first == t.Name {
		// well, I am part of the step, so search the tail in my children

		for _, c := range t.Childs {
			if c.Name == tail[0] {
				// this child might be part of the path, use it
				validChild = c
				break
			}
		}
	}
	if validChild == nil {
		// no child is capable of handling the tail
		// so I MUST BE the longest match, return my data
		return t.Fid, tail
	}
	fid, ctail := validChild.findLongestMatch(tail)
	if fid == nil {
		// I might have a valid child, but the child don't have a valid fid
		// this meas that we are just steps in a subpath that holds a valid mount
		// so return my data which is probably enough to walk to the desired
		// path
		return t.Fid, tail
	}
	// my child is at least one step deeper in the tree
	// so use its data
	return fid, ctail
}
