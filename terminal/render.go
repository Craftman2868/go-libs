package terminal

import (
	"os"
	"strconv"
)

const CSI = "\033["
const CR = "\r"
const LF = "\n"
const CRLF = CR + LF

const NEWLINE = CRLF

// Write

func Write(text string) (int, error) {
	return os.Stdout.Write([]byte(text))
}

func WriteBytes(bytes []byte) (int, error) {
	return os.Stdout.Write(bytes)
}

func Writeln(text string) (int, error) {
	n, err := Write(text)

	if err != nil {
		return n, err
	}

	n2, err2 := Write(NEWLINE)

	if err2 != nil {
		return n + n2, err2
	}

	return n + n2, nil
}

func WritelnBytes(text []byte) (int, error) {
	n, err := WriteBytes(text)

	if err != nil {
		return n, err
	}

	n2, err2 := Write(NEWLINE)

	if err2 != nil {
		return n + n2, err2
	}

	return n + n2, nil
}

func WriteAt(x, y int, text string) (int, error) {
	SetCursorPos(x, y)

	return Write(text)
}

func WriteBytesAt(x, y int, bytes []byte) (int, error) {
	SetCursorPos(x, y)

	return os.Stdout.Write(bytes)
}

// Mode

func EnableMode(mode string) {
	Write(CSI + "?" + mode + "h")
}

func DisableMode(mode string) {
	Write(CSI + "?" + mode + "l")
}

// Cursor

func SetCursorHome() {
	Write(CSI + "H")
}

func SetCursorPos(x, y int) {
	if x == 0 && y == 0 {
		SetCursorHome()
	}
	Write(CSI + strconv.Itoa(y+1) + ";" + strconv.Itoa(x+1) + "H")
}

func SaveCursor() {
	Write(CSI + "s")
}

func RestoreCursor() {
	Write(CSI + "u")
}

func HideCursor() {
	DisableMode("25")
}

func ShowCursor() {
	EnableMode("25")
}

// Clear

func ClearScreen() {
	Write(CSI + "2J")
}

func ClearHistory() {
	Write(CSI + "3J")
}

func ClearLine() {
	Write(CSI + "2K")
}

func ClearLineBegin() {
	Write(CSI + "1K")
}

func ClearLineEnd() {
	Write(CSI + "0K")
}

// Alternative screen

func EnableAltScreen() {
	EnableMode("1049")
}

func DisableAltScreen() {
	DisableMode("1049")
}

// Mouse tracking

func EnableMouseTracking() {
	EnableMode("1000")
	EnableMode("1005")
}

func DisableMouseTracking() {
	DisableMode("1005")
	DisableMode("1000")
}

// Style

func SetStyle(style string) {
	Write(CSI + style + "m")
}
