package eventdriven

// 事件循环组 用于持有事件循环的句柄对象
type eventPoolGroup struct {
	pools []*eventPool
}
type eventPool struct {
	poller *poller
	cs     map[int]*conn
}
