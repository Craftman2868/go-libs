package event

type Eventable interface {
	HandleEvent(event Event)
}
