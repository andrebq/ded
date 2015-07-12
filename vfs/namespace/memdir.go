package namespace

import (
	"errors"
)

type (
	// Memdir implements the Dir interface but uses a *Namespace
	// to hold the list of its contents.
	Memdir struct {
		ns Namespace
	}
)

func (m *Memdir) Readdir() ([]FileInfo, error) {
	m.ensureRoot()
	ret := make([]FileInfo, len(m.ns.root.childs))
	for i, n := range m.ns.root.childs {
		var err error
		ret[i], err = n.fileInfo()
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func (m *Memdir) Walk(n string) (File, error) {
	m.ensureRoot()
	if n, err := m.ns.root.Walk(n); err != nil {
		return nil, err
	} else {
		if !n.isLeaf() {
			// in our case, this is an error
			// since Memdir cannot have subdirs
			return nil, errors.New("path not found")
		}
		return n.file, nil
	}
}

func (m *Memdir) Stat() (FileInfo, error) {
	m.ensureRoot()
	return FileInfo{
		Name:  m.ns.root.name,
		IsDir: true,
	}, nil
}

func (m *Memdir) Seek(_ int64, _ int) (int64, error) {
	return 0, errors.New("cannot seek from memdir")
}

func (m *Memdir) Read(_ []byte) (int, error) {
	return 0, errors.New("cannot read from memdir")
}

func (m *Memdir) Write(_ []byte) (int, error) {
	return 0, errors.New("cannot write to memdir")
}

func (m *Memdir) Open() error {
	return nil
}

func (m *Memdir) Close() error {
	return nil
}

func (m *Memdir) AddFile(f File) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	err = m.ns.Mount(stat.Name, "", f)
	return err
}

func (m *Memdir) ensureRoot() {
	if m.ns.root == nil {
		m.ns.root = &node{}
	}
}
