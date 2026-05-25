package posture

import "time"

// WakeEvent represents a device wake event from sleep or hibernation, used to trigger
// posture re-evaluation when the system resumes from a suspended state.
type WakeEvent struct {
	At time.Time
}

// UnlockEvent represents a device unlock event after screen lock, used to trigger
// posture re-evaluation when user authentication state changes.
type UnlockEvent struct {
	At time.Time
}

// EventState provides platform-specific monitoring of system events that may affect
// posture compliance, such as waking from sleep or unlocking the device.
type EventState interface {
	// ListenForWake registers a callback for system wake events, returning a function
	// to stop listening.
	ListenForWake(func(WakeEvent)) (stop func(), err error)

	// ListenForUnlock registers a callback for device unlock events, returning a function
	// to stop listening.
	ListenForUnlock(func(event UnlockEvent)) (stop func(), err error)
}

var _ EventState = (*NoOpEventState)(nil)

// NoOpEventState is a placeholder implementation that stores callbacks without actually
// monitoring system events. Platform-specific implementations should be used for
// production deployments.
type NoOpEventState struct {
	onWake   func(WakeEvent)
	onUnlock func(UnlockEvent)
}

func (n *NoOpEventState) ListenForWake(f func(WakeEvent)) (stop func(), err error) {
	n.onWake = f
	return func() {
		n.onWake = nil
	}, nil
}

func (n *NoOpEventState) ListenForUnlock(f func(event UnlockEvent)) (stop func(), err error) {
	n.onUnlock = f
	return func() {
		n.onUnlock = nil
	}, nil
}
