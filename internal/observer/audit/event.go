package audit

type EventActionType string

const EventShortened = "shorten"
const EventFollow = "follow"

type Event struct {
	Timestamp int             `json:"ts"`
	Action    EventActionType `json:"action"`
	UserID    string          `json:"user_id"`
	URL       string          `json:"url"`
}
