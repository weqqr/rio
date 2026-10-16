//go:build !windows

package gpu

import "fmt"

func availableBackends() []Backend { return nil }

func newBackend(backend Backend) (instanceBackend, error) {
	return nil, fmt.Errorf("gpu: backend %s is not available on this platform", backend)
}
