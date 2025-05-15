package event

type Eventable struct {
	handlers map[string][]func(Event)
}

func (eventable *Eventable) HandleEvent(event Event) {
	if eventable.handlers == nil {
		return
	}

	for _, handler := range eventable.handlers[event.Name()] {
		handler(event)
	}
}

func (eventable *Eventable) On(eventName string, handler func(Event)) {
	if eventable.handlers == nil {
		eventable.handlers = make(map[string][]func(Event), 1)
	}
	eventable.handlers[eventName] = append(eventable.handlers[eventName], handler)
}
