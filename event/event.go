package event

type Event interface {
	Name() string
}

type BaseEvent struct {
	name string
}

func (ev BaseEvent) Name() string {
	return ev.name
}

func MakeEvent(name string) BaseEvent {
	return BaseEvent{name}
}
