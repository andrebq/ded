package vfs

import (
	"9fans.net/go/plan9"
	"errors"
	"io"
	"os"
)

type (
	File struct {
		Name    string
		content HasContent
		childs  []*File

		curfd FileContents
	}

	HasContent interface {
		Open(mode int, perm int) (FileContents, error)
		Size() (uint64, error)
		Close() error
	}

	FileContents interface {
		Read([]byte) (int, error)
		Write([]byte) (int, error)
		Seek(int64, SeekPoint) (int64, error)
		Sync() error
		Close() error
	}

	InMemory struct {
		buf []byte
		cur int
	}

	SeekPoint int
)

const (
	SeekStart   = SeekPoint(0)
	SeekCurrent = SeekPoint(1)
	SeekEnd     = SeekPoint(2)
)

func NewFile(name string, content HasContent) *File {
	return &File{
		Name:    name,
		childs:  make([]*File, 0),
		content: content,
	}
}

func (f *File) Size() (uint64, error) {
	if f.content == nil {
		return 0, nil
	}
	return f.content.Size()
}

func (f *File) Walk(name string) *File {
	for _, c := range f.childs {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (f *File) Contents() (FileContents, error) {
	if f.curfd != nil {
		return f.curfd, nil
	}
	return nil, errors.New("not open")
}

func (f *File) Open(mode int, perm int) (FileContents, error) {
	var err error
	if f.curfd != nil {
		return nil, errors.New("already open")
	}
	if f.content == nil {
		return nil, errors.New("content less")
	}
	f.curfd, err = f.content.Open(mode, perm)
	return f.curfd, err
}

func (f *File) Close() error {
	fd := f.curfd
	f.curfd = nil
	if fd == nil {
		return nil
	}
	return fd.Close()
}

func (f *File) Sync() error {
	fc, err := f.Contents()
	if err != nil {
		return err
	}
	return fc.Sync()
}

func (f *File) Add(c *File) {
	f.childs = append(f.childs, c)
}

func (in *InMemory) Open(mode int, perm int) (FileContents, error) {
	return in, nil
}

func (in *InMemory) Read(b []byte) (int, error) {
	if in.cur >= len(in.buf) {
		return 0, io.EOF
	}
	slice := in.buf[in.cur:]
	n := copy(b, slice)
	in.cur += n
	return n, nil
}

func (in *InMemory) Write(b []byte) (int, error) {
	in.Expand(len(b))

	// if we got here, then expand ensures
	// that b can be copied directly after in.cur
	buf := in.buf[in.cur : in.cur+len(b)]
	n := copy(buf, b)
	// move the cursor
	in.cur += n
	// expand the length to include the cursor
	in.buf = in.buf[0:in.cur]

	if n != len(b) {
		panic("len != n")
	}
	return n, nil
}

func (in *InMemory) Expand(sz int) {
	empty := cap(in.buf) - (in.cur + 1)
	if empty > sz {
		return
	}
	//TODO(andre): use a buffer cache
	nb := make([]byte, len(in.buf), len(in.buf)+sz)
	copy(nb, in.buf)
	in.buf = nb
	return
}

func (in *InMemory) Seek(sz int64, sp SeekPoint) (int64, error) {
	switch sp {
	case SeekStart:
		in.cur = int(sz)
		if in.cur > len(in.buf) {
			in.cur = len(in.buf)
		}
	case SeekEnd:
		in.cur = len(in.buf) - int(sz)
		if in.cur < 0 {
			in.cur = 0
		}
	case SeekCurrent:
		in.cur += int(sz)
		if in.cur > len(in.buf) {
			in.cur = len(in.buf)
		}
	}
	return int64(in.cur), nil
}

func (in *InMemory) Sync() error {
	return nil
}

func (in *InMemory) Close() error {
	//TODO(andre): recycle the buffer, but for now let the GC do the work
	in.buf = nil
	return nil
}

func (in *InMemory) Bytes() []byte {
	return in.buf
}

func (in *InMemory) Size() (uint64, error) {
	return uint64(len(in.buf)), nil
}

func FileInfoToDir(stat os.FileInfo) (dir plan9.Dir) {
	if stat.IsDir() {
		dir.Type |= plan9.QTDIR
	}
	dir.Atime = uint32(stat.ModTime().Unix())
	dir.Mtime = uint32(stat.ModTime().Unix())
	dir.Name = stat.Name()
	dir.Length = uint64(stat.Size())
	dir.Uid = "none"
	dir.Gid = "none"
	dir.Muid = "none"
	return
}

func DirToQid(dir plan9.Dir) plan9.Qid {
	return plan9.Qid{
		Path: 0,
		Vers: 1,
		Type: uint8(dir.Type),
	}
}

func DirModeToOSMode(dm uint32) (osmode int) {
	if (dm & plan9.OREAD) == plan9.OREAD {
		osmode |= os.O_RDONLY
	}
	if (dm & plan9.OWRITE) == plan9.OWRITE {
		osmode |= os.O_WRONLY
	}
	if (dm & plan9.ORDWR) == plan9.ORDWR {
		osmode |= os.O_RDWR
	}
	if (dm & plan9.OAPPEND) == plan9.OAPPEND {
		osmode |= os.O_APPEND
	}
	if (dm & plan9.OTRUNC) == plan9.OTRUNC {
		osmode |= os.O_TRUNC
	}
	return
}

func Plan9PermToUnix(p uint32) os.FileMode {
	return os.FileMode(p & 0777)
}
