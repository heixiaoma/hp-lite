package bean

import (
	"github.com/quic-go/quic-go"
	"net"
)

type WtoN struct {

	//通讯ID
	ChannelId string

	//外网通讯
	W quic.Stream

	//内网通讯
	N net.Conn
}
