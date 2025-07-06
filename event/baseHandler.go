package event

type BaseHandler struct {
	handlers map[string][]func(Event)
}

func (h *BaseHandler) HandleEvent(event Event) {
	if h.handlers == nil {
		return
	}

	for _, handler := range h.handlers[event.Name()] {
		handler(event)
	}
}

func (h *BaseHandler) On(eventName string, handler func(Event)) {
	if h.handlers == nil {
		h.handlers = make(map[string][]func(Event), 1)
	}

	h.handlers[eventName] = append(h.handlers[eventName], handler)
}

type onable interface {
	On(string, func(Event))
}

// Automatically bind the function to the matching event
// WARNING: MUST NOT be used with an interface as E (e.g. Event), E must be a struct that implement Event
func Bind[E Event](h onable, handler func(E)) {
	var proto E

	if any(proto) == nil {
		panic("E must not be an interface")
	}

	h.On(proto.Name(), func(ev Event) {
		handler(ev.(E))
	})
}
