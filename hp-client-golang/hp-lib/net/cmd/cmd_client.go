package cmd

import (
	net2 "hp-lib/net/hp"
)

var version = "hp-lite:6.0"

type CmdClient struct {
	CallMsg    func(message string)
	handler    *CmdClientHandler
	Connection *TcpConnection
}

func NewCmdClient(callMsg func(message string)) *CmdClient {
	return &CmdClient{
		CallMsg: callMsg,
	}
}

func (cmdClient *CmdClient) Connect(serverIp string, serverPort int, key string) {
	if cmdClient.Connection != nil {
		cmdClient.Connection.Close()
		cmdClient.Connection = nil
	}
	connection := NewTcpConnection()
	handler := &CmdClientHandler{
		Key:       key,
		CmdClient: cmdClient,
	}
	cmdClient.handler = handler
	connection.Connect(serverIp, serverPort, handler, cmdClient.CallMsg)
	cmdClient.Connection = connection
}

func (cmdClient *CmdClient) GetStatus() bool {
	if cmdClient.handler != nil {
		if !cmdClient.handler.Ide() {
			return false
		}
		return true
	} else {
		return false
	}
}

func (cmdClient *CmdClient) Close() {
	net2.CloseTunnel()
	if cmdClient.Connection != nil {
		cmdClient.Connection.Close()
		cmdClient.Connection = nil
	}
}
