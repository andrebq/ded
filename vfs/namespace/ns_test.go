package namespace

import (
	"bytes"
	"testing"
)

func TestMount(t *testing.T) {
	ns := NewNamespace()

	mdir := Memdir{}

	// make an empty node
	if err := ns.Mount("mem", "", nil); err != nil {
		t.Fatal(err)
	}

	if err := ns.Mount("mdir", "/mem", &mdir); err != nil {
		t.Fatal(err)
	}

	if fr, err := ns.Open("/mem/mdir"); err != nil {
		t.Fatal(err)
	} else {
		if fr.AbsPath != "/mem/mdir" {
			t.Errorf("invalid abspath. should be %v got %v", "/mem/mdir", fr.AbsPath)
		}
		if fr.File != (&mdir) {
			t.Errorf("Invalid file reference. Should be %v got %v", mdir, fr.File)
		}
	}

	rdonly := NewReadOnly("a.readonly", []byte(`hello`))
	wronly := NewWriteOnly("a.writeonly")
	mdir.AddFile(rdonly)
	mdir.AddFile(wronly)

	if fr, err := ns.Open("/mem/mdir/a.readonly"); err != nil {
		t.Fatal(err)
	} else {
		tmp := make([]byte, 5)
		if n, err := fr.Read(tmp); err != nil {
			t.Fatal(err)
		} else {
			if n != len(tmp) {
				t.Errorf("short read. should be %v got %v", len(tmp), n)
			} else if !bytes.Equal(tmp, rdonly.readBuf) {
				t.Errorf("read mismatch")
			}
		}
	}

	if fr, err := ns.Open("/mem/mdir/a.writeonly"); err != nil {
		t.Fatal(err)
	} else {
		tmp := []byte(`hello`)
		if n, err := fr.Write(tmp); err != nil {
			t.Fatal(err)
		} else {
			if n != len(tmp) {
				t.Errorf("short write. should be %v got %v", len(tmp), n)
			} else if !bytes.Equal(tmp, wronly.Bytes()) {
				t.Errorf("write mismatch")
			}
		}
	}
}
