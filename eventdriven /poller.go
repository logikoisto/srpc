package eventdriven

import (
	"golang.org/x/sys/unix"
	"log"
)

// Poller represents a poller which is in charge of monitoring file-descriptors.
type poller struct {
	fd int // epoll fd
}

// openPoller instantiates a poller.
func openPoller() (pr *poller, err error) {
	pr = new(poller)
	if pr.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); err != nil {
		pr = nil
		return
	}
	return
}

// Close closes the poller.
func (p *poller) close() error {
	return unix.Close(p.fd)
}

// Polling blocks the current goroutine, waiting for network-events.
func (p *poller) polling(callback func(fd int, ev uint32) error) (err error) {
	el := newEventList(InitEvents)
	for {
		n, err0 := unix.EpollWait(p.fd, el.events, -1)
		if err0 != nil && err0 != unix.EINTR {
			log.Println(err0)
			continue
		}
		for i := 0; i < n; i++ {
			fd := int(el.events[i].Fd)
			if err = callback(fd, el.events[i].Events); err != nil {
				return
			}
		}
		if n == el.size {
			el.increase()
		}
	}
}

const (
	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
	writeEvents     = unix.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
)

// AddReadWrite registers the given file-descriptor with readable and writable events to the poller.
func (p *poller) addReadWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readWriteEvents})
}

// AddRead registers the given file-descriptor with readable event to the poller.
func (p *poller) addRead(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readEvents})
}

// AddWrite registers the given file-descriptor with writable event to the poller.
func (p *poller) addWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Fd: int32(fd), Events: writeEvents})
}

// ModRead renews the given file-descriptor with readable event in the poller.
func (p *poller) modRead(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readEvents})
}

// ModReadWrite renews the given file-descriptor with readable and writable events in the poller.
func (p *poller) modReadWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{Fd: int32(fd), Events: readWriteEvents})
}

// Delete removes the given file-descriptor from the poller.
func (p *poller) delete(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd, nil)
}
