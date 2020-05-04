package server

import (
	"net"
	"srpc/eventdriven "
)

func InitRPC(addr *net.TCPAddr) {
	eventdriven.StartServer(addr, func(conn eventdriven.Connection, action string, act func()) (lastData []byte, err error) {

		return []byte{},nil
	})
}
