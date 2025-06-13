package app

import (
	"github.com/Craftman2868/go-libs/clock"
	"github.com/Craftman2868/go-libs/event"
)

type App struct {
	event.BaseHandler
	clock.Clock

	running bool
}

func (app *App) Init(tps int) {
	app.running = false
	app.Clock = clock.NewClock(tps)

	app.HandleEvent(InitEvent{})
}

func (app *App) Quit() {
	app.HandleEvent(QuitEvent{})
}

func (app *App) Update() {
	app.HandleEvent(UpdateEvent{})
}

func (app *App) Run() {
	app.running = true

	for app.running {
		app.Update()
		app.TickSleep()
	}
}

func (app *App) IsRunning() bool {
	return app.running
}

func (app *App) Stop() {
	app.running = false
}

func NewApp() App {
	return App{}
}
