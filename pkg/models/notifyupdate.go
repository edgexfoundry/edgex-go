package models

const (
	NotifyUpdateAdd    = "add"
	NotifyUpdateUpdate = "update"
	NotifyUpdateDelete = "delete"
)

type NotifyUpdate struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
}

