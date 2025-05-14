package tunnel

import (
	"bufio"
	"errors"
	"github.com/pires/go-proxyproto"
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
	conn      *base.MuxSession
	stream    *base.MuxStream
	channelId string
	userInfo  bean.UserConfigInfo
}

func NewTcpHandler(tcpConn net.Conn, conn *base.MuxSession, userInfo bean.UserConfigInfo) *TcpHandler {
	return &TcpHandler{conn: conn, channelId: util.NewId(), tcpConn: tcpConn, userInfo: userInfo}
}

func (h *TcpHandler) handlerStream(stream *base.MuxStream) {
	defer func() {
		if stream.IsTcp {
			stream.TcpStream.Close()
		} else {
			stream.QuicStream.Close()
		}
	}()

	var reader *bufio.Reader
	if stream.IsTcp {
		reader = bufio.NewReader(stream.TcpStream)
	} else {
		reader = bufio.NewReader(stream.QuicStream)
	}
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
		if receiver.stream.IsTcp {
			receiver.stream.TcpStream.Close()
		} else {
			receiver.stream.QuicStream.Close()
		}
	}
}

// ChannelActive 连接激活时，发送注册信息给云端
func (h *TcpHandler) ChannelActive(conn net.Conn) {
	var stream *base.MuxStream
	var err error
	if h.conn.IsTcp {
		stream1, err1 := h.conn.TcpSession.OpenStream()
		err = err1
		stream = &base.MuxStream{IsTcp: true, TcpStream: stream1}

	} else {
		stream2, err2 := h.conn.QuicSession.OpenStream()
		err = err2
		stream = &base.MuxStream{IsTcp: false, QuicStream: stream2}
	}
	if err == nil {
		h.stream = stream
		m := &message.HpMessage{
			Type: message.HpMessage_CONNECTED,
			MetaData: &message.HpMessage_MetaData{
				Type:      message.HpMessage_TCP,
				ChannelId: h.channelId,
			},
		}

		if stream.IsTcp {
			stream.TcpStream.Write(protol.Encode(m))
		} else {
			stream.QuicStream.Write(protol.Encode(m))
		}

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
					if stream.IsTcp {
						stream.TcpStream.Write(protol.Encode(m))
					} else {
						stream.QuicStream.Write(protol.Encode(m))
					}
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
	if h.stream.IsTcp {
		h.stream.TcpStream.Write(protol.Encode(m))
	} else {
		h.stream.QuicStream.Write(protol.Encode(m))
	}
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
		if h.stream.IsTcp {
			h.stream.TcpStream.Write(protol.Encode(m))
			h.stream.TcpStream.Close()
		} else {
			h.stream.QuicStream.Write(protol.Encode(m))
			h.stream.QuicStream.Close()
		}

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
