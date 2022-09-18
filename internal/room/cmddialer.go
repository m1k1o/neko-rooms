package room

import (
	"fmt"
	"io"
	"net"
	"time"
)

type cmdDialer struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (c *cmdDialer) Read(b []byte) (n int, err error) {
	n, err = c.stdout.Read(b)
	fmt.Printf("Read: %s\n", b)
	return
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (c *cmdDialer) Write(b []byte) (n int, err error) {
	fmt.Printf("Write: %s\n", b)
	return c.stdin.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *cmdDialer) Close() error {
	if err := c.stdin.Close(); err != nil {
		return err
	}
	return c.stdout.Close()
}

// LocalAddr returns the local network address, if known.
func (c *cmdDialer) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns the remote network address, if known.
func (c *cmdDialer) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *cmdDialer) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *cmdDialer) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *cmdDialer) SetWriteDeadline(t time.Time) error {
	return nil
}
