package models

// Constants for AdminState
const (
	// Locked : device is locked
	// Unlocked : device is unlocked
	Locked   = "LOCKED"
	Unlocked = "UNLOCKED"
)

// Constants for ChannelType
const (
	Rest  = "REST"
	Email = "EMAIL"
)

// Constants for NotificationSeverity
const (
	Minor    = "MINOR"
	Critical = "CRITICAL"
	Normal   = "NORMAL"
)

// Constants for NotificationStatus
const (
	New       = "NEW"
	Processed = "PROCESSED"

	EscalationSubscriptionName = "ESCALATION"
	EscalationPrefix           = "escalated-"
	EscalatedContentNotice     = "This notification is escalated by the transmission"
)

// Constants for TransmissionStatus
const (
	Failed       = "FAILED"
	Sent         = "SENT"
	Acknowledged = "ACKNOWLEDGED"
	RESENDING    = "RESENDING"
)

// Constants for both NotificationStatus and TransmissionStatus
const (
	Escalated = "ESCALATED"
)

// Constants for OperatingState
const (
	Up      = "UP"
	Down    = "DOWN"
	Unknown = "UNKNOWN"
)
