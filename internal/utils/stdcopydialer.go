package utils

import (
	"io"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

type stdCopyDialer struct {
	net.Conn
	stdout io.ReadCloser
}

func StdCopyDialer(res types.HijackedResponse) net.Conn {
	stdout, dstout := io.Pipe()
	go stdcopy.StdCopy(dstout, io.Discard, res.Conn)

	return &stdCopyDialer{
		Conn:   res.Conn,
		stdout: stdout,
	}
}

func (d *stdCopyDialer) Read(b []byte) (int, error) {
	return d.stdout.Read(b)
}
