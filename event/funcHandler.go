package event

type FuncHandler func(event Event)

func (f FuncHandler) HandleEvent(event Event) {
	f(event)
}
