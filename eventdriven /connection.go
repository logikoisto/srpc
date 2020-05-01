package eventdriven

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

// 缓冲区对象
type buf struct {
	bf []byte
}

func (b *buf) IsEmpty() bool {
	return true
}
