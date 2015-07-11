package namespace

import (
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

	if fr, err := ns.Walk("/mem/mdir"); err != nil {
		t.Fatal(err)
	} else {
		if fr.AbsPath != "/mem/mdir" {
			t.Errorf("invalid abspath. should be %v got %v", "/mem/mdir", fr.AbsPath)
		}
		if fr.File != (&mdir) {
			t.Errorf("Invalid file reference. Should be %v got %v", mdir, fr.File)
		}
	}
}
