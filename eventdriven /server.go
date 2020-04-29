package eventdriven

import (
	"net"
	"runtime"
)

// server rpc服务的对外对象
type server struct {
	fd    int
	ln    *net.TCPListener
	epg   *eventPoolGroup
	clean *cleaner
}

// 构建器
func StartServer(addr *net.TCPAddr) {
	var (
		err error
		ln  *net.TCPListener
	)
	if ln, err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}
	f, _ := ln.File()
	svr := server{
		fd:    int(f.Fd()),
		ln:    ln,
		clean: &cleaner{shutDownChan: make(chan struct{})},
	}
	eg := &eventPoolGroup{
		pools: make([]*eventPool, runtime.NumCPU()),
	}
	for i := 0; i < len(eg.pools); i++ {
		if eg.pools[i].poller, err = openPoller(); err != nil {
			panic(err)
		}
		// TODO: 这里需要对 连接的注册进行 封装 -> eventPool
		eg.pools[i].cs = make(map[int]*conn, 0)
		if err = eg.pools[i].poller.AddRead(svr.fd); err != nil {
			panic(err)
		}
		eg.pools[i].cs[svr.fd] = &conn{}
	}
	// TODO:启动 server
	svr.run()
}

// 资源清理通知器
type cleaner struct {
	shutDownChan chan struct{}
}

func (c *cleaner) shutDown() {
	close(c.shutDownChan)
}

func (srv *server) run() {
}
