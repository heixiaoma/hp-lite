package bean

type ConnectType string

const (
	TCP     ConnectType = "TCP"
	UDP     ConnectType = "UDP"
	TCP_UDP ConnectType = "TCP_UDP"
)
