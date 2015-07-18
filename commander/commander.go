package commander

import (
	"bytes"
	"errors"
	"github.com/flynn/go-shlex"
	"gopkg.in/pipe.v2"
	"io"
)

type (
	C struct {
	}
)

// RunLine blocks and execute the command at line.
func (c *C) RunLine(line string) ([]byte, error) {
	args, err := shlex.Split(line)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, errors.New("no command to run")
	}

	cmd := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = nil
	}

	p := pipe.Exec(cmd, args...)
	return pipe.Output(p)
}

func (c *C) RunWithInputBytes(line string, in []byte) ([]byte, error) {
	return c.RunWithInput(line, bytes.NewReader(in))
}

func (c *C) RunWithInput(line string, in io.Reader) ([]byte, error) {
	args, err := shlex.Split(line)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, errors.New("no command to run")
	}

	cmd := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = nil
	}

	p := pipe.Line(
		pipe.Read(in),
		pipe.Exec(cmd, args...),
	)
	return pipe.Output(p)
}
