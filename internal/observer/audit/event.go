package audit

// EventActionType представляет тип действия в событии аудита.
type EventActionType string

// EventShortened - константа для действия "сокращение URL".
const EventShortened = "shorten"

// EventFollow - константа для действия "переход по URL".
const EventFollow = "follow"

// Event представляет событие аудита в системе.
// Содержит информацию о времени события, типе действия, пользователе и URL.
type Event struct {
	Timestamp int             `json:"ts"`
	Action    EventActionType `json:"action"`
	UserID    string          `json:"user_id"`
	URL       string          `json:"url"`
}
