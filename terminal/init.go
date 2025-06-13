package terminal

// Init / quit

func Init() {
	InitInput()

	HideCursor()
	SaveCursor()
	EnableAltScreen()
	EnableMouseTracking()

	ClearScreen()
	SetCursorHome()
}

func Quit() {
	ClearScreen()
	DisableMouseTracking()
	DisableAltScreen()
	RestoreCursor()
	ShowCursor()

	QuitInput()
}
