package terminal

// Init / quit

func Init() {
	InitInput()

	HideCursor()
	EnableAltScreen()
	EnableMouseTracking()

	SetCursorHome()
}

func Quit() {
	DisableMouseTracking()
	DisableAltScreen()
	ShowCursor()

	QuitInput()
}
