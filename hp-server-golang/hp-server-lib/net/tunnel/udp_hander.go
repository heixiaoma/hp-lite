package tunnel

import (
	"bufio"
	"hp-server-lib/message"
	"hp-server-lib/net/base"
	"hp-server-lib/protol"
	"hp-server-lib/util"
	"log"
	"net"
	"time"
)

type UdpHandler struct {
	udpConn      *net.UDPConn
	conn         *base.MuxSession
	stream       *base.MuxStream
	addr         *net.UDPAddr
	channelId    string
	udpServer    *UdpServer
	lastActiveAt time.Time
}

func NewUdpHandler(udpServer *UdpServer, udpConn *net.UDPConn, conn *base.MuxSession, addr *net.UDPAddr) *UdpHandler {
	return &UdpHandler{udpServer: udpServer, udpConn: udpConn, conn: conn, channelId: util.NewId(), addr: addr, lastActiveAt: time.Now()}
}
func (h *UdpHandler) handlerStream(stream *base.MuxStream) {
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

func (receiver *UdpHandler) ReadStreamData(data *message.HpMessage) {
	if data.Type == message.HpMessage_DATA {
		log.Printf(string(data.Data))
		receiver.lastActiveAt = time.Now()
		receiver.udpConn.WriteToUDP(data.Data, receiver.addr)
	}
	if data.Type == message.HpMessage_DISCONNECTED {
		receiver.udpConn.Close()
		if receiver.stream.IsTcp {
			receiver.stream.TcpStream.Close()
		} else {
			receiver.stream.QuicStream.Close()
		}
	}
}

func (h *UdpHandler) ChannelActive(udpConn *net.UDPConn) {
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
		m := &message.HpMessage{
			Type: message.HpMessage_CONNECTED,
			MetaData: &message.HpMessage_MetaData{
				Type:      message.HpMessage_UDP,
				ChannelId: h.channelId,
			},
		}
		if stream.IsTcp {
			stream.TcpStream.Write(protol.Encode(m))
		} else {
			stream.QuicStream.Write(protol.Encode(m))
		}
		log.Printf("通知内网连接")
		h.stream = stream
		go h.handlerStream(stream)
	}
	go func() {
		// 创建一个每 5 秒触发一次的定时器
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop() // 确保定时器最终被停止
		// 无限循环，每 5 秒执行一次任务
		for {
			select {
			case <-ticker.C:
				sub := time.Now().Sub(h.lastActiveAt)
				if sub.Seconds() > 20 {
					value, ok := h.udpServer.cache.Load(h.addr.String())
					if ok {
						handler := value.(*UdpHandler)
						handler.ChannelInactive(h.udpConn)
						h.udpServer.cache.Delete(h.channelId)
						return
					}
				}
			}
		}
	}()
}

func (h *UdpHandler) ChannelRead(udpConn *net.UDPConn, data interface{}) {
	m := &message.HpMessage{
		Type: message.HpMessage_DATA,
		MetaData: &message.HpMessage_MetaData{
			Type:      message.HpMessage_UDP,
			ChannelId: h.channelId,
		},
		Data: data.([]byte),
	}
	if h.stream != nil {
		if h.stream.IsTcp {
			h.stream.TcpStream.Write(protol.Encode(m))
		} else {
			h.stream.QuicStream.Write(protol.Encode(m))
		}
		h.lastActiveAt = time.Now()
	}
}

func (h *UdpHandler) ChannelInactive(udpConn *net.UDPConn) {
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
