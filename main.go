package main

import (
	"github.com/weqqr/rio/pkg/renderer"
	"github.com/weqqr/rio/pkg/wev"
)

type App struct {
	renderer *render.Renderer
}

func (a *App) OnWindowCreated(e *wev.EventLoop, window *wev.Window) error {
	renderer, err := render.New(window.RawWindowHandle())
	if err != nil {
		return err
	}
	a.renderer = renderer
	return a.renderer.Render()
}

func (a *App) OnWindowResize(e *wev.EventLoop, window *wev.Window, width, height int) error {
	if a.renderer == nil || width == 0 || height == 0 {
		return nil
	}
	if err := a.renderer.Resize(uint32(width), uint32(height)); err != nil {
		return err
	}
	return a.renderer.Render()
}

func (a *App) OnCloseRequested(e *wev.EventLoop, window *wev.Window) error {
	e.Quit()
	return nil
}

func main() {
	eventLoop := wev.NewEventLoop()

	app := &App{}
	_, err := eventLoop.NewWindow(wev.WindowConfig{Title: "rio"})
	if err != nil {
		panic(err)
	}

	if err := eventLoop.Run(app); err != nil {
		panic(err)
	}
	if app.renderer != nil {
		if err := app.renderer.Close(); err != nil {
			panic(err)
		}
	}
	if err := eventLoop.Close(); err != nil {
		panic(err)
	}
}
