package ext

type ForwardProxy interface {
	Start() bool
	Stop()
}
