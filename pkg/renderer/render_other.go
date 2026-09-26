//go:build !windows

package render

import "errors"

type Renderer struct{}

func New(rawWindowHandle uintptr) (*Renderer, error) {
	return nil, errors.New("render: unsupported platform")
}

func (r *Renderer) Render() error { return errors.New("render: unsupported platform") }

func (r *Renderer) Resize(width, height uint32) error {
	return errors.New("render: unsupported platform")
}

func (r *Renderer) Close() error { return nil }
