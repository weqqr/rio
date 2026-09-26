//go:build !windows

package wev

func NewEventLoop() *EventLoop {
	panic("wev: unsupported platform")
}
