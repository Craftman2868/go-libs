package event

type Event interface {
	Name() string
}

type BaseEvent string

func (ev BaseEvent) Name() string {
	return string(ev)
}
