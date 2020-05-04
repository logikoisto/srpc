package eventdriven

const (
	READ = "read"
)

// 暴露在外 供用户层使用的 事件处理器
type EventHandler func(conn Connection, action string, act func()) (lastData []byte, err error)
