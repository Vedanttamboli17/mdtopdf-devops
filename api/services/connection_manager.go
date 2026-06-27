package services

import (
	"sync"

	"github.com/docplatform/shared/metrics"
)

// JobEvent is pushed to SSE clients when a job reaches a terminal state.
type JobEvent struct {
	FileID       string `json:"file_id"`
	Status       string `json:"status"`
	DownloadURL  string `json:"download_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// ConnectionManager tracks active SSE connections keyed by file_id.
// Each file_id maps to a bidirectional chan JobEvent (cap 1) so events
// arriving between Register() and the stream writer starting are not dropped.
type ConnectionManager struct {
	m sync.Map // map[string]chan JobEvent
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{}
}

// Register creates a notification channel for fileID and increments the active
// SSE gauge. If a channel already exists (e.g. duplicate browser tab), the old
// one is closed and the gauge is decremented first.
func (cm *ConnectionManager) Register(fileID string) chan JobEvent {
	ch := make(chan JobEvent, 1)
	if old, ok := cm.m.LoadAndDelete(fileID); ok {
		close(old.(chan JobEvent))
		metrics.ActiveSSEConnections.Dec()
	}
	cm.m.Store(fileID, ch)
	metrics.ActiveSSEConnections.Inc()
	return ch
}

// Notify delivers an event to the registered connection for fileID, removes the
// registration, and decrements the gauge. No-op if no connection is registered.
func (cm *ConnectionManager) Notify(fileID string, event JobEvent) {
	if raw, ok := cm.m.LoadAndDelete(fileID); ok {
		ch := raw.(chan JobEvent)
		select {
		case ch <- event:
		default:
		}
		metrics.ActiveSSEConnections.Dec()
	}
}

// Deregister removes the entry for fileID only if ch is still the registered
// channel (guards the race where a newer Register() superseded ours before we
// got here). Decrements the gauge on removal.
func (cm *ConnectionManager) Deregister(fileID string, ch chan JobEvent) {
	if cm.m.CompareAndDelete(fileID, ch) {
		metrics.ActiveSSEConnections.Dec()
	}
}
