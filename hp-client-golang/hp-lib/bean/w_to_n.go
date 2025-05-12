package bean

import (
	"github.com/hashicorp/yamux"
	"net"
)

type WtoN struct {

	//通讯ID
	ChannelId string

	//外网通讯
	W *yamux.Stream

	//内网通讯
	N net.Conn
}
