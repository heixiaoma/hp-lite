package tunnel

import (
	"hp-server-lib/bean"
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
	userInfo     bean.UserConfigInfo
	protocol     string
	localIp      string
	localPort    int
}

func NewUdpHandler(udpServer *UdpServer, udpConn *net.UDPConn, conn *base.MuxSession, addr *net.UDPAddr, userInfo bean.UserConfigInfo) (error, *UdpHandler) {
	err, s, s2, i := util.ProtocolInfo(userInfo.LocalAddress)
	if err != nil {
		return err, nil
	}
	return nil, &UdpHandler{udpServer: udpServer, udpConn: udpConn, conn: conn, channelId: util.NewId(), addr: addr, lastActiveAt: time.Now(), protocol: s, localIp: s2, localPort: i}
}
func (h *UdpHandler) handlerStream(stream *base.MuxStream) {
	defer stream.Close()
	reader := stream.GetReader()
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
		receiver.lastActiveAt = time.Now()
		receiver.udpConn.WriteToUDP(data.Data, receiver.addr)
	}
	if data.Type == message.HpMessage_DISCONNECTED {
		receiver.udpConn.Close()
		receiver.stream.Close()
	}
}

func (h *UdpHandler) ChannelActive(udpConn *net.UDPConn) {
	stream, err := h.conn.OpenStream()
	if err == nil {
		m := &message.HpMessage{
			Type: message.HpMessage_CONNECTED,
			MetaData: &message.HpMessage_MetaData{
				Protocol:  h.protocol,
				ChannelId: h.channelId,
			},
		}
		stream.Write(protol.Encode(m))
		log.Printf("通知内网连接")
		h.stream = stream
		go h.handlerStream(stream)
	} else {
		log.Println("UDP服务激活创建流失败:" + err.Error())
		err := h.conn.Close()
		if err != nil {
			log.Println("UDP服务关闭失败:" + err.Error())
		}
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
				if sub.Seconds() > 60 {
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
			Protocol:  h.protocol,
			ChannelId: h.channelId,
		},
		Data: data.([]byte),
	}
	if h.stream != nil {
		h.stream.Write(protol.Encode(m))
		h.lastActiveAt = time.Now()
	}
}

func (h *UdpHandler) ChannelInactive(udpConn *net.UDPConn) {
	m := &message.HpMessage{
		Type: message.HpMessage_DISCONNECTED,
		MetaData: &message.HpMessage_MetaData{
			Protocol:  h.protocol,
			ChannelId: h.channelId,
		},
	}
	if h.stream != nil {
		h.stream.Write(protol.Encode(m))
		h.stream.Close()
	}
}
