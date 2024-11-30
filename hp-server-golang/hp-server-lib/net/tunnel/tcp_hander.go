package tunnel

import (
	"bufio"
	"github.com/quic-go/quic-go"
	"hp-server-lib/message"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"io"
	"net"
)

type TcpHandler struct {
	tcpConn   net.Conn
	conn      quic.Connection
	stream    quic.Stream
	channelId string
}

func NewTcpHandler(tcpConn net.Conn, conn quic.Connection) *TcpHandler {
	return &TcpHandler{conn: conn, channelId: util.NewId(), tcpConn: tcpConn}
}

func (h *TcpHandler) handlerStream(stream quic.Stream) {
	defer stream.Close()
	reader := bufio.NewReader(stream)
	//避坑点：多包问题，需要重复读取解包
	for {
		decode, e := protol.Decode(reader)
		if e != nil {
			return
		}
		if decode != nil {
			h.ReadStreamData(decode)
		}
	}
}

func (receiver *TcpHandler) ReadStreamData(data *message.HpMessage) {
	if data.Type == message.HpMessage_DATA {
		receiver.tcpConn.Write(data.Data)
	}
	if data.Type == message.HpMessage_DISCONNECTED {
		receiver.tcpConn.Close()
		receiver.stream.Close()
	}
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *TcpHandler) ChannelActive(conn net.Conn) {
	stream, err := h.conn.OpenStream()
	if err == nil {
		m := &message.HpMessage{
			Type: message.HpMessage_CONNECTED,
			MetaData: &message.HpMessage_MetaData{
				Type:      message.HpMessage_TCP,
				ChannelId: h.channelId,
			},
		}
		stream.Write(protol.Encode(m))
		h.stream = stream
		go h.handlerStream(stream)
	}
}

func (h *TcpHandler) ChannelRead(conn net.Conn, data interface{}) {
	m := &message.HpMessage{
		Type: message.HpMessage_DATA,
		MetaData: &message.HpMessage_MetaData{
			Type:      message.HpMessage_TCP,
			ChannelId: h.channelId,
		},
		Data: data.([]byte),
	}
	h.stream.Write(protol.Encode(m))
}

func (h *TcpHandler) ChannelInactive(conn net.Conn) {
	m := &message.HpMessage{
		Type: message.HpMessage_DISCONNECTED,
		MetaData: &message.HpMessage_MetaData{
			Type:      message.HpMessage_TCP,
			ChannelId: h.channelId,
		},
	}
	h.stream.Write(protol.Encode(m))
	h.stream.Close()
}

func (h *TcpHandler) Decode(reader *bufio.Reader) (interface{}, error) {
	if reader.Buffered() > 0 {
		data := make([]byte, reader.Buffered())
		_, err := io.ReadFull(reader, data)
		return data, err
	}
	return nil, nil
}
