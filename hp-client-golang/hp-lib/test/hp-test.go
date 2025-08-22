package main

import (
	"hp-lib/bean"
	"hp-lib/net/hp"
	"time"
)

func main() {

	client := hp.NewHpClient(func(message string) {
		println(message)
	})

	for true {
		client.Connect(&bean.LocalInnerWear{
			OutLimit:    -1,
			InLimit:     -1,
			Md5:         "",
			ConnectType: "TCP",
			ConfigKey:   "ac9b0b97-b3dd-44ef-9570-0a0039823398",
			LocalIp:     "192.168.1.1",
			LocalPort:   80,
			ServerIp:    "47.243.162.173",
			//ServerIp:   "127.0.0.1",
			ServerPort: 9091,
		})
		time.Sleep(1 * time.Second)
		//client.Close()
	}

}
