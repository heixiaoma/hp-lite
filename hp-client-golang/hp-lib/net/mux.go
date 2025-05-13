package net

import (
	"github.com/quic-go/quic-go"
	"github.com/xtaci/smux"
)

type MuxStream struct {
	IsTcp      bool
	QuicStream quic.Stream
	TcpStream  *smux.Stream
}

type MuxSession struct {
	IsTcp       bool
	QuicSession quic.Connection
	TcpSession  *smux.Session
}
