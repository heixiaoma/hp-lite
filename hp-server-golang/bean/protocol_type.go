package bean

type Protocol string

const (
	HTTP    Protocol = "http"
	HTTPS   Protocol = "https"
	TCP     Protocol = "tcp"
	UDP     Protocol = "udp"
	SOCKS5  Protocol = "socks5"
	UNIX    Protocol = "unix"
	TCP_UDP Protocol = "tcp_udp"
)
