package eventdriven

// 暴露在外 供用户层使用的 事件处理器
type EventHandler func(data []byte, action string) []byte
