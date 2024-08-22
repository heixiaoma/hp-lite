package cmd

import (
	net2 "hp-lib/net/hp"
	"net"
)

var version = "hp-pro:5.0"

type CmdClient struct {
	CallMsg func(message string)
	conn    net.Conn
	handler *CmdClientHandler
}

func NewCmdClient(callMsg func(message string)) *CmdClient {
	return &CmdClient{
		CallMsg: callMsg,
	}
}

func (cmdClient *CmdClient) Connect(serverIp string, serverPort int, key string) {
	if cmdClient.conn != nil {
		cmdClient.conn.Close()
	}
	connection := NewTcpConnection()
	handler := &CmdClientHandler{
		Key:       key,
		CmdClient: cmdClient,
	}
	cmdClient.handler = handler
	cmdClient.conn = connection.Connect(serverIp, serverPort, handler, cmdClient.CallMsg)
}

func (cmdClient *CmdClient) GetStatus() bool {
	if cmdClient.handler != nil {
		return cmdClient.handler.Active
	} else {
		return false
	}
}

func (cmdClient *CmdClient) Close() {
	net2.CloseTunnel()
	if cmdClient.conn != nil {
		cmdClient.conn.Close()
	}
}
