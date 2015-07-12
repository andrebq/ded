package namespace

import (
	"errors"
	"path"
	"strings"
)

type (
	Namespace struct {
		root *node
	}

	node struct {
		name   string
		childs []*node
		file   File
		leaf   bool
	}

	File interface {
		Open() error
		Close() error
		Stat() (FileInfo, error)
		Seek(offset int64, whence int) (int64, error)
		Read(out []byte) (int, error)
		Write(in []byte) (int, error)
		Readdir() ([]FileInfo, error)
		Walk(n string) (File, error)
		Sync() error
	}

	FileRef struct {
		File
		AbsPath string
	}

	FileInfo struct {
		Name  string
		IsDir bool
		Size  uint64
	}
)

func NewNamespace() *Namespace {
	return &Namespace{
		root: &node{},
	}
}

/** Mount is used to expose file under name at dir.

File can be nil, in this case there name will just be saved as a tree node
with no data.
*/
func (n *Namespace) Mount(name string, dir string, file File) error {
	if n.root == nil {
		return errors.New("namespace not ready")
	}
	parts := path.Clean(dir)
	root := n.root
	var err error
	for _, p := range strings.Split(parts, "/") {
		root, err = root.Walk(p)
		if err != nil {
			return err
		}
	}

	_, err = root.AddChild(name, file)
	return err
}

/** Walk search the namespace for the path p and returns the file associated with
the path (if any) or an error indicating that the path couldn't be found.

The returned reference holds the absolute path used to reach it.
*/
func (n *Namespace) Open(p string) (FileRef, error) {
	if n.root == nil {
		return FileRef{}, errors.New("namespace not ready")
	}
	p = path.Clean(p)
	root := n.root
	var err error
	for _, part := range strings.Split(p, "/") {
		root, err = root.Open(part)
		if err != nil {
			return FileRef{}, err
		}
	}

	err = root.file.Open()
	if err != nil {
		return FileRef{}, err
	}
	return FileRef{
		File:    root.file,
		AbsPath: p,
	}, nil
}

/** Walk checks if name is a child of n and returns it.

If n is a node pointing to a virtual directory, then that directory,
is searched for name.

After a successfull search into a virtual directory, the returned file,
is added as a child node of n. Furter calls to Walk will see that node instead of
searching the virtual directory. */
func (n *node) Walk(name string) (*node, error) {
	if len(name) == 0 || name == "." {
		return n, nil
	}
	for _, c := range n.childs {
		if c.name == name {
			return c, nil
		}
	}

	if n.file == nil {
		return nil, errors.New("path not found")
	}

	f, err := n.file.Walk(name)
	if err != nil {
		return nil, err
	}

	return n.AddChild(name, f)
}

func (n *node) Open(name string) (*node, error) {
	if len(name) == 0 || name == "." {
		return n, nil
	}
	for _, c := range n.childs {
		if c.name == name {
			return c, nil
		}
	}

	if n.file == nil {
		return nil, errors.New("path not found")
	}

	f, err := n.file.Walk(name)
	if err != nil {
		return nil, err
	}

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &node{
		name: name,
		file: f,
		leaf: st.IsDir,
	}, nil
}

/** AddChild exposes f as a child of n with the given name.

If the name is already used, the old one is replaced with the new file, but only
if there is no virtual file or if the virtual file is a virtual directory.

The virtual directory (if any) isn't touched, the requirement exists just to avoid
add a child file to a leaf node.*/
func (n *node) AddChild(name string, f File) (*node, error) {
	if n.isLeaf() {
		return nil, errors.New("cannot add child to a leaf")
	}

	nnode := &node{
		name: name,
		file: f,
	}
	if f == nil {
		nnode.leaf = false
	} else {
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}
		nnode.leaf = !stat.IsDir
	}
	for i, c := range n.childs {
		if c.name == name {
			n.childs[i] = nnode
			return nnode, nil
		}
	}
	// not a replacement, add new
	n.childs = append(n.childs, nnode)
	return nnode, nil
}

// check if the node points to a virtual file
func (n *node) isLeaf() bool {
	return n.leaf
}

// fileInfo returns some file information about this node
func (n *node) fileInfo() (FileInfo, error) {
	if n.file != nil {
		return n.file.Stat()
	}
	return FileInfo{
		Name: n.name,
		// by default a node not pointing to a file is a directory
		IsDir: true,
	}, nil
}
