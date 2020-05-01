package eventdriven

import (
	"golang.org/x/sys/unix"
)

// 事件循环组 用于持有事件循环的句柄对象
type eventPoolGroup struct {
	pools []*eventPool
}
type eventPool struct {
	serverFD int
	poller   *poller
	cs       map[int]*conn
}

func createEventPool(fd int) *eventPool {
	if p, err := openPoller(); err != nil {
		panic(err)
	} else {
		return &eventPool{
			serverFD: fd,
			poller:   p,
			cs:       make(map[int]*conn, 0),
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

// TODO: 事件处理
func (ep *eventPool) handleEvent(fd int, ev uint32) error {
	if c, ok := ep.cs[fd]; ok {
		if !c.outBuf.IsEmpty() {
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
			// TODO: 增加连接数 & 开启连接
			// return ep.loopOpen(c)
		}
		return err
	}
	return nil
}

func (ep *eventPool) write(c *conn) error {
	return nil
}
func (ep *eventPool) read(c *conn) error {
	return nil
}
