package vfs

import (
	"bytes"
	"testing"
)

func TestInMemory(t *testing.T) {
	in := &InMemory{}
	if n, err := in.Write([]byte(`hello`)); err != nil {
		t.Fatalf("Error writing: %v", err)
	} else if n != 5 {
		t.Error("Short write")
	}

	if _, err := in.Seek(0, SeekStart); err != nil {
		t.Fatalf("Error on seek: %v", err)
	}

	aux := make([]byte, 5)
	if n, err := in.Read(aux); err != nil {
		t.Fatalf("Invalid read %v", err)
	} else if n != 5 {
		t.Fatalf("Short read")
	} else if !bytes.Equal(aux, []byte(`hello`)) {
		t.Fatalf("Wrong contents: %v", string(aux))
	}

	if n, err := in.Write([]byte(`hello`)); err != nil {
		t.Fatalf("Error writing: %v", err)
	} else if n != 5 {
		t.Error("Short write")
	}

	if _, err := in.Seek(0, SeekStart); err != nil {
		t.Fatalf("Error on seek: %v", err)
	}

	aux = make([]byte, 10)
	if n, err := in.Read(aux); err != nil {
		t.Fatalf("Invalid read %v", err)
	} else if n != 10 {
		t.Fatalf("Short read")
	} else if !bytes.Equal(aux, []byte(`hellohello`)) {
		t.Fatalf("Wrong contents: %v", string(aux))
	}

	if _, err := in.Seek(5, SeekEnd); err != nil {
		t.Fatalf("Error on seek: %v", err)
	}

	aux = make([]byte, 5)
	if n, err := in.Read(aux); err != nil {
		t.Fatalf("Invalid read %v", err)
	} else if n != 5 {
		t.Fatalf("Short read")
	} else if !bytes.Equal(aux, []byte(`hello`)) {
		t.Fatalf("Wrong contents: %v", string(aux))
	}

	if _, err := in.Seek(5, SeekStart); err != nil {
		t.Fatalf("Error on seek: %v", err)
	}
	if _, err := in.Seek(1, SeekCurrent); err != nil {
		t.Fatalf("Error on seek: %v", err)
	}

	aux = make([]byte, 4)
	if n, err := in.Read(aux); err != nil {
		t.Fatalf("Invalid read %v", err)
	} else if n != 4 {
		t.Fatalf("Short read")
	} else if !bytes.Equal(aux, []byte(`ello`)) {
		t.Fatalf("Wrong contents: %v", string(aux))
	}
}
