package app

import (
	"github.com/Craftman2868/go-libs/event"
	"github.com/Craftman2868/go-libs/terminal"
)

type TerminalApp struct {
	App
	parser terminal.Parser
}

func (app *TerminalApp) InitTerminal() {
	app.On("init", func(event.Event) {
		terminal.Init()
		app.parser = terminal.NewParser(app)
	})

	app.On("quit", func(event.Event) {
		terminal.Quit()
	})

	app.On("update", func(event.Event) {
		app.parser.HandleInputs()
	})
}
