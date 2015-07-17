package buffer

import (
	"bytes"
	"testing"
)

func TestBuffer(t *testing.T) {
	b := &B{}
	if n, e := b.Write([]byte("hello")); e != nil {
		t.Fatal(e)
	} else if n != 5 {
		t.Errorf("invalid write")
	}

	if !bytes.Equal(b.Bytes(), []byte("hello")) {
		t.Errorf("wrong data returned")
	}

	b.Seek(0, 0)

	if n, e := b.Write([]byte("olleh")); e != nil {
		t.Fatal(e)
	} else if n != 5 {
		t.Errorf("invalid write")
	}

	if !bytes.Equal(b.Bytes(), []byte("olleh")) {
		t.Errorf("wrong data returned")
	}
}
