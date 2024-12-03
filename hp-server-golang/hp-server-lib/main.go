package main

import "hp-server-lib/web"

func main() {

	//go http.StartHttpServer()
	//go http.StartHttpsServer()
	//
	//tcpServer := server.NewCmdServer(server.NewCmdHandler())
	//go tcpServer.StartServer(9091)
	//
	//quicServer := server.NewHPServer(server.NewHPHandler())
	//go quicServer.StartServer(9090)
	go web.StartWebServer(9090)
	println("---")
	select {}
}
