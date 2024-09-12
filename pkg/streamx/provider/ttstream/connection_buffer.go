package ttstream

import (
	"github.com/cloudwego/gopkg/bufiox"
	"github.com/cloudwego/netpoll"
)

var _ bufiox.Reader = (*connBuffer)(nil)
var _ bufiox.Writer = (*connBuffer)(nil)

func newConnBuffer(conn netpoll.Connection) *connBuffer {
	return &connBuffer{conn: conn}
}

type connBuffer struct {
	conn      netpoll.Connection
	readSize  int
	writeSize int
}

func (c *connBuffer) Next(n int) (p []byte, err error) {
	p, err = c.conn.Reader().Next(n)
	c.readSize += len(p)
	return p, err
}

func (c *connBuffer) ReadBinary(bs []byte) (n int, err error) {
	n = len(bs)
	buf, err := c.conn.Reader().Next(n)
	if err != nil {
		return 0, err
	}
	copy(bs, buf)
	c.readSize += n
	return n, nil
}

func (c *connBuffer) Peek(n int) (buf []byte, err error) {
	return c.conn.Reader().Peek(n)
}

func (c *connBuffer) Skip(n int) (err error) {
	err = c.conn.Reader().Skip(n)
	if err != nil {
		return err
	}
	c.readSize += n
	return nil
}

func (c *connBuffer) ReadLen() (n int) {
	return c.readSize
}

func (c *connBuffer) Release(e error) (err error) {
	c.readSize = 0
	return c.conn.Reader().Release()
}

func (c *connBuffer) Malloc(n int) (buf []byte, err error) {
	c.writeSize += n
	return c.conn.Writer().Malloc(n)
}

func (c *connBuffer) WriteBinary(bs []byte) (n int, err error) {
	n, err = c.conn.Writer().WriteBinary(bs)
	c.writeSize += n
	return n, err
}

func (c *connBuffer) WrittenLen() (length int) {
	return c.writeSize
}

func (c *connBuffer) Flush() (err error) {
	c.writeSize = 0
	return c.conn.Writer().Flush()
}
