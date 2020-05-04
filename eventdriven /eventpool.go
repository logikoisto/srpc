package eventdriven

import (
	"golang.org/x/sys/unix"
	"sync/atomic"
)

// 事件循环组 用于持有事件循环的句柄对象
type eventPoolGroup struct {
	pools []*eventPool
}
type eventPool struct {
	serverFD int
	poller   *poller
	cs       map[int]*conn
	connNum  int32
	packet   []byte
	callback EventHandler
}

func createEventPool(fd int, eventHandler EventHandler) *eventPool {
	if p, err := openPoller(); err != nil {
		panic(err)
	} else {
		return &eventPool{
			serverFD: fd,
			poller:   p,
			cs:       make(map[int]*conn, 0),
			packet:   make([]byte, 1024),
			callback: eventHandler,
		}
	}
}
func (ep *eventPool) registerFD(fd int) {
	if err := ep.poller.addRead(fd); err != nil {
		panic(err)
	}
	ep.cs[fd] = initConn(fd)
}

func (ep *eventPool) run() {
	if err := ep.poller.polling(ep.handleEvent); err != nil {
		panic(err)
	}
}

func (ep *eventPool) removeFD(fd int) {
	if err := ep.poller.delete(fd); err != nil {
		panic(err)
	}
	delete(ep.cs, fd)
}
func (ep *eventPool) addConnNum() {
	atomic.AddInt32(&ep.connNum, 1)
}
func (ep *eventPool) delConnNum() {
	atomic.AddInt32(&ep.connNum, -1)
}

func (ep *eventPool) closeConn(c *conn, err error) error {
	c.close()
	return err
}

func (ep *eventPool) handleEvent(fd int, ev uint32) error {
	if c, ok := ep.cs[fd]; ok {
		if !c.outBuf.isEmpty() {
			if ev&OutEvents != 0 {
				return ep.write(c)
			}
			return nil
		} else {
			if ev&InEvents != 0 {
				return ep.read(c)
			}
			return nil
		}
	}
	return ep.accept(fd)
}

func (ep *eventPool) accept(fd int) error {
	if ep.serverFD == fd {
		nfd, sa, err := unix.Accept(fd)
		if err != nil {
			if err == unix.EAGAIN {
				return nil
			}
			return err
		}
		if err = unix.SetNonblock(nfd, true); err != nil {
			return err
		}
		_ = sa
		// c := newConn(nfd, el, sa)
		c := initConn(nfd)
		if err = ep.poller.addRead(c.fd); err == nil {
			ep.cs[c.fd] = c
			// TODO: 开启连接
			// return ep.loopOpen(c)
			ep.addConnNum()
		}
		return err
	}
	return nil
}

func (ep *eventPool) write(c *conn) error {
	head := c.outBuf.readAll()
	n, err := unix.Write(c.fd, head)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return ep.closeConn(c, err)
	}
	c.outBuf.shift(n)
	if c.outBuf.isEmpty() {
		_ = ep.poller.modRead(c.fd)
	}
	return nil
}

func (ep *eventPool) read(c *conn) error {
	n, err := unix.Read(c.fd, ep.packet)
	if n == 0 || err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return ep.closeConn(c, err)
	}
	c.inBuf.write(ep.packet[:n])
	var lastData []byte
	if lastData, err = ep.callback(Connection{c: c}, READ, func() {
		if !c.outBuf.isEmpty() {
			_ = ep.poller.modReadWrite(c.fd)
		}
	}); err != nil {
		return err
	}
	// 这样做 是为了在回调函数中处理半包问题
	if len(lastData) != 0 {
		c.inBuf.write(lastData)
	}
	return nil
}
