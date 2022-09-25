package bufio

import (
	"io"
	"net"
	"time"

	"github.com/Dreamacro/clash/common/buf"
)

// copy from https://github.com/SagerNet/sing/blob/1bb95f9661fc2a8f2a31886db2fd04a4da75abc5/common/bufio/cache.go

type CachedConn struct {
	net.Conn
	buffer *buf.Buffer
}

func NewCachedConn(conn net.Conn, buffer *buf.Buffer) *CachedConn {
	return &CachedConn{
		Conn:   conn,
		buffer: buffer,
	}
}

func (c *CachedConn) ReadCached() *buf.Buffer {
	buffer := c.buffer
	c.buffer = nil
	return buffer
}

func (c *CachedConn) Read(p []byte) (n int, err error) {
	if c.buffer != nil {
		n, err = c.buffer.Read(p)
		if err == nil {
			if c.buffer.IsEmpty() {
				cN, err := c.Conn.Read(p[n:])
				if err == io.EOF {
					err = nil
				}
				return n + cN, err

			}
			return
		}

		c.buffer.Release()
		c.buffer = nil
	}

	return c.Conn.Read(p)
}

//func (c *CachedConn) WriteTo(w io.Writer) (n int64, err error) {
//	if c.buffer != nil {
//		wn, wErr := w.Write(c.buffer.Bytes())
//		if wErr != nil {
//			c.buffer.Release()
//			c.buffer = nil
//		}
//		n += int64(wn)
//	}
//	cn, err := Copy(w, c.Conn)
//	n += cn
//	return
//}

func (c *CachedConn) SetReadDeadline(t time.Time) error {
	if c.buffer != nil && !c.buffer.IsEmpty() {
		return nil
	}
	return c.Conn.SetReadDeadline(t)
}

//func (c *CachedConn) ReadFrom(r io.Reader) (n int64, err error) {
//	return Copy(c.Conn, r)
//}

func (c *CachedConn) Upstream() any {
	return c.Conn
}

func (c *CachedConn) ReaderReplaceable() bool {
	return c.buffer == nil
}

func (c *CachedConn) WriterReplaceable() bool {
	return true
}

func (c *CachedConn) Close() error {
	c.buffer.Release()
	return c.Conn.Close()
}
