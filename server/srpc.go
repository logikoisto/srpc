package server

import (
	"net"
	"srpc/eventdriven "
)

func InitRPC(addr *net.TCPAddr) {
	eventdriven.StartServer(addr, func(conn eventdriven.Connection, action string, act func()) (lastData []byte, err error) {
		// TODO: 实现  读取 编码 应用逻辑 解码 发送 等逻辑
		return []byte{},nil
	})
}
