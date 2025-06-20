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
