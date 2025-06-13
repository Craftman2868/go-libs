package app

type InitEvent struct{}

func (ev InitEvent) Name() string {
	return "init"
}

type QuitEvent struct{}

func (ev QuitEvent) Name() string {
	return "quit"
}

type UpdateEvent struct{}

func (ev UpdateEvent) Name() string {
	return "update"
}
