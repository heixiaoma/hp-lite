package main

import (
	"hp-server-lib/net/server"
)

func main() {
	tcpServer := server.NewCmdServer(server.NewCmdHandler())
	go tcpServer.StartServer(9091)

	quicServer := server.NewHPServer(server.NewHPHandler())
	go quicServer.StartServer(9090)
	println("---")
	select {}
}
