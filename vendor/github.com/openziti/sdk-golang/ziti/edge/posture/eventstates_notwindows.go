//go:build !windows

package posture

// NewEventState is a stand-in for actual non-Windows event watching
func NewEventState() EventState {
	return &NoOpEventState{}
}
