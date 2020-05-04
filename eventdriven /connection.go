package eventdriven

type Connection struct {
	c *conn
}

// 连接对象
type conn struct {
	fd     int // socket 文件描述符
	inBuf  buf
	outBuf buf
}

func initConn(fd int) *conn {
	return &conn{
		fd:     fd,
		inBuf:  buf{make([]byte, 0)},
		outBuf: buf{make([]byte, 0)},
	}
}

func (c *Connection) ReadData() []byte {
	return c.c.inBuf.readAll()
}
func (c *Connection) WriteData(data []byte) {
	c.c.outBuf.write(data)
}

// TODO: 释放资源
func (c *conn) close() {
}

// 缓冲区对象
type buf struct {
	bf []byte
}

func (b *buf) shift(n int) {
	if n >= len(b.bf) {
		return
	}
	b.bf = b.bf[n:]
}

func (b *buf) readAll() []byte {
	return b.bf
}
func (b *buf) write(in []byte) {
	b.bf = append(b.bf, in...)
}

func (b *buf) isEmpty() bool {
	return true
}
