package main

import "github.com/weqqr/rio/pkg/wev"

type App struct{}

func (a *App) OnCloseRequested(e *wev.EventLoop, window *wev.Window) error {
	e.Quit()
	return nil
}

func main() {
	eventLoop := wev.NewEventLoop()

	_, err := eventLoop.NewWindow(wev.WindowConfig{Title: "rio"})
	if err != nil {
		panic(err)
	}

	if err := eventLoop.Run(&App{}); err != nil {
		panic(err)
	}
	if err := eventLoop.Close(); err != nil {
		panic(err)
	}
}
