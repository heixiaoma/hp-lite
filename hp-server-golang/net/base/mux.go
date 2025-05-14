package base

import (
	"bufio"
	"github.com/quic-go/quic-go"
	"github.com/xtaci/smux"
	"net"
)

type MuxStream struct {
	isTcp      bool
	QuicStream quic.Stream
	TcpStream  *smux.Stream
}

func (receiver *MuxStream) GetReader() *bufio.Reader {
	var reader *bufio.Reader
	if receiver.isTcp && receiver.TcpStream != nil {
		reader = bufio.NewReader(receiver.TcpStream)
	} else if receiver.QuicStream != nil {
		reader = bufio.NewReader(receiver.QuicStream)
	}
	return reader
}

func (receiver *MuxStream) Write(b []byte) (n int, err error) {
	if receiver.isTcp && receiver.TcpStream != nil {
		return receiver.TcpStream.Write(b)
	} else if receiver.QuicStream != nil {
		return receiver.QuicStream.Write(b)
	}
	return 0, err
}

func (receiver *MuxStream) StreamID() interface{} {
	if receiver.isTcp && receiver.TcpStream != nil {
		return receiver.TcpStream.ID()
	} else if receiver.QuicStream != nil {
		return receiver.QuicStream.StreamID()
	}
	return nil
}

func (receiver *MuxStream) Close() error {
	if receiver.isTcp && receiver.TcpStream != nil {
		return receiver.TcpStream.Close()
	} else if receiver.QuicStream != nil {
		return receiver.QuicStream.Close()
	}
	return nil
}

func NewQuicMuxStream(stream quic.Stream) *MuxStream {
	return &MuxStream{isTcp: false, QuicStream: stream}
}

func NewTcpMuxStream(stream *smux.Stream) *MuxStream {
	return &MuxStream{isTcp: true, TcpStream: stream}
}

type MuxSession struct {
	isTcp       bool
	QuicSession quic.Connection
	TcpSession  *smux.Session
}

func NewTcpMuxSession(session *smux.Session) *MuxSession {
	return &MuxSession{isTcp: true, TcpSession: session}
}

func NewQuicMuxSession(session quic.Connection) *MuxSession {
	return &MuxSession{isTcp: false, QuicSession: session}
}

func (receiver *MuxSession) OpenStream() (*MuxStream, error) {
	var stream *MuxStream
	var err error
	if receiver.isTcp {
		stream1, err1 := receiver.TcpSession.OpenStream()
		err = err1
		stream = &MuxStream{isTcp: true, TcpStream: stream1}

	} else {
		stream2, err2 := receiver.QuicSession.OpenStream()
		err = err2
		stream = &MuxStream{isTcp: false, QuicStream: stream2}
	}
	return stream, err
}

func (receiver *MuxSession) Close() error {
	if receiver.isTcp && receiver.TcpSession != nil {
		return receiver.TcpSession.Close()
	} else if receiver.QuicSession != nil {
		return receiver.QuicSession.CloseWithError(0, "正常关闭")
	}
	return nil
}

func (receiver *MuxSession) RemoteAddr() net.Addr {
	if receiver.isTcp && receiver.TcpSession != nil {
		return receiver.TcpSession.RemoteAddr()
	} else if receiver.QuicSession != nil {
		return receiver.QuicSession.RemoteAddr()
	}
	return nil
}
