package models

const (
	EventTypeImageUpdate = "ImageUpdate"
)

type Event struct {
	EventType string
	Time      string
	Desc      string
}
