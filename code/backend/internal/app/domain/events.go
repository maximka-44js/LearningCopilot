package domain

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type BaseEvent struct {
	at time.Time
}

func NewBaseEvent() BaseEvent {
	return BaseEvent{at: time.Now().UTC()}
}

func (e BaseEvent) OccurredAt() time.Time {
	return e.at
}
