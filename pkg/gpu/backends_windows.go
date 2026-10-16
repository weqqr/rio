package gpu

import "fmt"

func availableBackends() []Backend { return []Backend{BackendD3D12} }

func newBackend(backend Backend) (instanceBackend, error) {
	switch backend {
	case BackendD3D12:
		return newD3D12Instance()
	case BackendVulkan:
		return nil, fmt.Errorf("gpu: backend %s is not implemented", backend)
	}
	return nil, fmt.Errorf("gpu: unknown backend %d", int(backend))
}
