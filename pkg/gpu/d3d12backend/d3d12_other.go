//go:build !windows

package d3d12backend

import (
	"errors"

	"github.com/weqqr/rio/pkg/gpu"
)

func New() (gpu.InstanceBackend, error) {
	return nil, errors.New("gpu: d3d12 backend is not available on this platform")
}
