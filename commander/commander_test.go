package commander

import (
	"bytes"
	"testing"
)

func TestCommander(t *testing.T) {
	// assume we have an echo command
	c := &C{}
	if out, err := c.RunLine("echo 'abc123'"); err != nil {
		t.Fatalf("Error reading command output: %v", err)
	} else if !bytes.Equal(out, []byte("abc123\n")) {
		t.Errorf("Invalid output. %v", string(out))
	}
}

func TestCommanderWithInput(t *testing.T) {
	c := &C{}
	in := bytes.NewBufferString("hello")
	if out, err := c.RunWithInput("cat -", in); err != nil {
		t.Fatalf("Error reading command output: %v", err)
	} else if !bytes.Equal(out, []byte("hello")) {
		t.Errorf("Invalid output: %v", string(out))
	}
}
