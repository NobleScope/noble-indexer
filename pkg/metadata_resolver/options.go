package metadata_resolver

import "time"

type ModuleOption func(*Module)

func WithSyncPeriod(t time.Duration) ModuleOption {
	return func(m *Module) {
		m.syncPeriod = t
	}
}
