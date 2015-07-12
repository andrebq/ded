package namespace

import (
	"errors"
)

type (
	Memwriter struct {
		buf []byte
		pos int
	}
)

func (me *Memwriter) Write(in []byte) (int, error) {
	me.expand(len(in))
	n := copy(me.buf[me.pos:], in)
	me.pos += n
	return n, nil
}

func (me *Memwriter) Seek(offset int64, whence int) (int64, error) {
	if whence != 0 {
		return 0, errors.New("seek is valid only from the begining of the buffer")
	}
	if int(offset) > len(me.buf) {
		return 0, errors.New("seek is larger than buffer")
	}
	me.pos = int(offset)
	return int64(me.pos), nil
}

func (me *Memwriter) Bytes() []byte {
	return me.buf
}

// expand ensures that it will be possible to copy sz bytes starting
// at the current position without the need to allocation another buffer
func (me *Memwriter) expand(sz int) {
	//TODO(andre) avoid some allocations
	if me.pos+sz > len(me.buf) {
		nbuf := make([]byte, me.pos+sz)
		copy(nbuf, me.buf)
		me.buf = nbuf
	}
}
