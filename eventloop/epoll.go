package multiplex

import (
	"golang.org/x/sys/unix"
	"net"
	"reflect"
	"runtime"
	"sync"
	"syscall"
)

type EPollConf struct {
	EpSize        int
	L             *net.TCPListener
	WaitQueueSize int
}

type ProcFuc func(conn *net.TCPConn) []byte

func Init(ef *EPollConf) {
	setLimit()
	pool{
		done:  make(chan struct{}),
		eSize: ef.EpSize,
		ln:    ef.L,
	}.run()
}

type pool struct {
	eSize int
	done  chan struct{}
	c     *EPollConf
	ln    *net.TCPListener
	f     ProcFuc
}

func (ep *pool) run() {
	for i := 0; i < runtime.NumCPU(); i++ {
		e, _ := newEpoll()
		// 创建与cpu 核数相同的 Accept 协程 发挥多核能力
		go func() {
			for {
				select {
				case <-ep.done:
					break
				default:
				}
				conn, err := ep.ln.AcceptTCP()
				setTcpConfig(conn)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Temporary() {
						continue
					}
				}
				if err := e.add(conn); err != nil {
					_ = conn.Close() //登录未成功直接关闭连接
					continue
				}
			}
		}()
		// 按配置数量创建 epoll wait 协程序 处理异步的读取与写入
		for i := 0; i < ep.eSize; i++ {
			go func() {
				for {
					select {
					case <-ep.done:
						return
					default:
						connections, err := e.wait()
						if err != nil {
							continue
						}
						for _, conn := range connections {
							if conn == nil {
								break
							}
							_ = ep.f(conn)
							//TODO: 处理异步写回请求
						}
					}
				}
			}()
		}
	}
}

func (ep *pool) Close() error {
	close(ep.done)
	return nil
}

func setTcpConfig(c *net.TCPConn) {
	_ = c.SetKeepAlive(true)
}

// epoll 操作对象
type epoll struct {
	fd            int
	WaitQueueSize int
	connections   sync.Map //TODO:这里适合使用 sync.map 吗?
}

func newEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:            fd,
		connections:   sync.Map{},
		WaitQueueSize: ep.c.WaitQueueSize,
	}, nil
}

// TODO: 默认水平触发模式,可采用非阻塞FD,优化边沿触发模式
func (e *epoll) add(conn *net.TCPConn) error {
	// Extract file descriptor associated with the connection
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	e.connections.Store(fd, conn)
	return nil
}
func (e *epoll) remove(conn *net.TCPConn) error {
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.connections.Delete(fd)
	return nil
}
func (e *epoll) wait() ([]*net.TCPConn, error) {
	events := make([]unix.EpollEvent, e.WaitQueueSize)
	n, err := unix.EpollWait(e.fd, events, e.WaitQueueSize)
	if err != nil {
		return nil, err
	}
	var connections []*net.TCPConn
	for i := 0; i < n; i++ {
		//log.Printf("event:%+v\n", events[i])
		if conn, ok := e.connections.Load(int(events[i].Fd)); ok {
			connections = append(connections, conn.(*net.TCPConn))
		}
	}
	return connections, nil
}
func socketFD(conn *net.TCPConn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(*conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

// 设置go 进程打开文件数的限制
func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
}
