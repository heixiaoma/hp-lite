package forward

type ForwardProxy interface {
	Start(close func()) bool
	Stop()
}
