package sshutil

import "io"

var _ io.Reader = (*nullReader)(nil)

type nullReader struct {
}

// Read implements io.Reader.
func (*nullReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func NewNullReader() io.Reader {
	return &nullReader{}
}
