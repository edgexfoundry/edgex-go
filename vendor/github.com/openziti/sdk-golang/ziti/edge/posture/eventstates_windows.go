//go:build windows

package posture

// NewEventState is a stand-in for actual Window event watching
func NewEventState() EventState {
	return &NoOpEventState{}
}
