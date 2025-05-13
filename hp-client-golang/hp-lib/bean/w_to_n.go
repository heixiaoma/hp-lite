package bean

import (
	"github.com/xtaci/smux"
	"net"
)

type WtoN struct {

	//通讯ID
	ChannelId string

	//外网通讯
	W *smux.Stream

	//内网通讯
	N net.Conn
}
