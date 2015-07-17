package buffer

import (
	"errors"
)

type B struct {
	buf []byte
	pos int
	sz  int
}

func (b *B) Bytes() []byte {
	return b.buf[:b.sz]
}

func (b *B) Write(in []byte) (int, error) {
	b.expand(len(in))
	sz := copy(b.buf[b.pos:], in)
	b.sz += sz
	return sz, nil
}

func (b *B) Seek(offset int64, whence int) (n int64, err error) {
	if whence != 0 {
		return 0, errors.New("seek only from the beginning")
	}
	if offset > int64(b.sz) {
		offset = int64(b.sz)
	}
	b.pos = int(offset)
	b.sz = b.pos
	return int64(b.pos), nil
}

func (b *B) expand(sz int) {
	if b.pos+sz <= len(b.buf) {
		// fine, buf is large enough
		return
	}
	tmp := make([]byte, b.pos+sz)
	copy(tmp, b.buf)
	b.buf = tmp
}
