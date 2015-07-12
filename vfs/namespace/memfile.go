package namespace

import (
	"bytes"
	"errors"
	"io"
)

type (
	Memfile struct {
		FileInfo
		readBuf []byte
		reader  *bytes.Reader
		writer  *Memwriter
	}
)

func NewReadOnly(name string, in []byte) *Memfile {
	return &Memfile{
		FileInfo: FileInfo{
			Name:  name,
			IsDir: false,
		},
		readBuf: in,
	}
}

func NewWriteOnly(name string) *Memfile {
	return &Memfile{
		FileInfo: FileInfo{
			Name:  name,
			IsDir: false,
		},
	}
}

func (f *Memfile) Stat() (FileInfo, error) {
	if f.readBuf != nil {
		f.FileInfo.Size = uint64(len(f.readBuf))
	} else if f.writer != nil {
		f.FileInfo.Size = uint64(len(f.writer.buf))
	}
	return f.FileInfo, nil
}

func (f *Memfile) Seek(offset int64, whence int) (int64, error) {
	if whence != 0 {
		return 0, errors.New("seek only from starting of file")
	}

	switch {
	case f.writer != nil:
		return f.writer.Seek(offset, whence)
	case f.reader != nil:
		return f.reader.Seek(offset, whence)
	}
	return 0, errors.New("invalid state")
}

func (f *Memfile) Read(out []byte) (int, error) {
	if f.reader == nil {
		return 0, io.EOF
	}
	return f.reader.Read(out)
}

func (f *Memfile) Write(in []byte) (int, error) {
	if f.writer == nil {
		return 0, io.EOF
	}
	return f.writer.Write(in)
}

func (f *Memfile) Open() error {
	if len(f.readBuf) > 0 {
		f.reader = bytes.NewReader(f.readBuf)
	} else {
		f.writer = &Memwriter{}
	}
	return nil
}

func (f *Memfile) Close() error {
	f.reader = nil
	return nil
}

func (f *Memfile) Readdir() ([]FileInfo, error) {
	return nil, errors.New("not a directory")
}

func (f *Memfile) Walk(_ string) (File, error) {
	return nil, errors.New("not a directory")
}

func (m *Memfile) Sync() error {
	if m.writer == nil {
		return errors.New("file is readonly")
	}
	return nil
}

func (f *Memfile) Bytes() []byte {
	if f.writer == nil {
		return nil
	}
	return f.writer.Bytes()
}
