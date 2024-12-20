package tunnel

import (
	"bufio"
	"errors"
	"github.com/pires/go-proxyproto"
	"github.com/quic-go/quic-go"
	"hp-server-lib/bean"
	"hp-server-lib/message"
	"hp-server-lib/net/base"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"io"
	"net"
	"strings"
)

type TcpHandler struct {
	tcpConn   net.Conn
	conn      quic.Connection
	stream    quic.Stream
	channelId string
	userInfo  bean.UserConfigInfo
}

func NewTcpHandler(tcpConn net.Conn, conn quic.Connection, userInfo bean.UserConfigInfo) *TcpHandler {
	return &TcpHandler{conn: conn, channelId: util.NewId(), tcpConn: tcpConn, userInfo: userInfo}
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
		base.AddSent(receiver.userInfo.ConfigId, int64(len(data.Data)))
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
		h.stream = stream
		m := &message.HpMessage{
			Type: message.HpMessage_CONNECTED,
			MetaData: &message.HpMessage_MetaData{
				Type:      message.HpMessage_TCP,
				ChannelId: h.channelId,
			},
		}
		stream.Write(protol.Encode(m))

		//检查是否开启了v1 v2代理,ton //&&!strings.Contains(conn.RemoteAddr().String(),"127.0.0.1")
		if len(h.userInfo.ProxyVersion) > 0 {
			var version byte = 0
			if strings.Compare(h.userInfo.ProxyVersion, "V1") == 0 {
				version = 1
			}
			if strings.Compare(h.userInfo.ProxyVersion, "V2") == 0 {
				version = 2
			}
			if version > 0 {
				header := &proxyproto.Header{
					Version:           version,
					Command:           proxyproto.PROXY,
					TransportProtocol: proxyproto.TCPv4,
					SourceAddr:        conn.RemoteAddr(),
					DestinationAddr: &net.TCPAddr{
						IP:   net.ParseIP(h.userInfo.ProxyIp),
						Port: h.userInfo.ProxyPort,
					},
				}
				format, err := header.Format()
				if err == nil {
					m := &message.HpMessage{
						Type: message.HpMessage_DATA,
						MetaData: &message.HpMessage_MetaData{
							Type:      message.HpMessage_TCP,
							ChannelId: h.channelId,
						},
						Data: format,
					}
					stream.Write(protol.Encode(m))
				}
			}
		}

		go h.handlerStream(stream)
	}
}

func (h *TcpHandler) ChannelRead(conn net.Conn, data interface{}) error {
	m := &message.HpMessage{
		Type: message.HpMessage_DATA,
		MetaData: &message.HpMessage_MetaData{
			Type:      message.HpMessage_TCP,
			ChannelId: h.channelId,
		},
		Data: data.([]byte),
	}
	base.AddReceived(h.userInfo.ConfigId, int64(len(m.Data)))
	if h.stream == nil {
		return errors.New("流关闭异常")
	}
	h.stream.Write(protol.Encode(m))
	return nil
}

func (h *TcpHandler) ChannelInactive(conn net.Conn) {
	m := &message.HpMessage{
		Type: message.HpMessage_DISCONNECTED,
		MetaData: &message.HpMessage_MetaData{
			Type:      message.HpMessage_TCP,
			ChannelId: h.channelId,
		},
	}
	if h.stream != nil {
		h.stream.Write(protol.Encode(m))
		h.stream.Close()
	}
}

func (h *TcpHandler) Decode(reader *bufio.Reader) (interface{}, error) {
	if reader.Buffered() > 0 {
		data := make([]byte, reader.Buffered())
		_, err := io.ReadFull(reader, data)
		return data, err
	}
	return nil, nil
}
