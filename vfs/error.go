package vfs

import (
	"9fans.net/go/plan9"
	"errors"
)

var (
	ErrInvalidFid = errors.New("invalid fid")
)

func PackError(fc *plan9.Fcall, err error) *plan9.Fcall {
	fc.Type = plan9.Rerror
	fc.Ename = err.Error()
	return fc
}
