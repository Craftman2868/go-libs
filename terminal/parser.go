package terminal

import (
	"strconv"
	"time"

	"github.com/Craftman2868/go-libs/event"
)

const (
	CSI_ARGV_BUFFER_SIZE = 8
	ALT_TIMEOUT          = 50 * time.Millisecond
)

const ESC = 27

type Parser struct {
	handler event.Eventable

	escape      bool
	escape_time time.Time

	csi      bool
	csi_argv [CSI_ARGV_BUFFER_SIZE]uint8
	csi_argc uint8

	ss3 bool
}

// Events

type KeyEvent struct {
	Ch  byte
	Key byte
	Mod uint8
}

func (ev KeyEvent) Name() string {
	return "key"
}

type MouseEvent struct {
	Button uint8
	Mod    uint8
	X, Y   uint8
}

func (ev MouseEvent) Name() string {
	return "mouse"
}

const ( // MouseEvent.Button
	BUTTON1 = iota
	BUTTON2
	BUTTON3
	RELEASE
)

const ( // MouseEvent.Mod / KeyEvent.Mod
	MOD_SHIFT = 1 << (iota + 2)
	MOD_META
	MOD_CTRL

	MOD_ALT = MOD_META
)

const ( // Special keys
	KEY_NULL = 0

	// CSI ... ~
	KEY_INSERT   = 2
	KEY_DELETE   = 3
	KEY_PAGEUP   = 5
	KEY_PAGEDOWN = 6
	KEY_F1       = 11
	KEY_F2       = 12
	KEY_F3       = 13
	KEY_F4       = 14
	KEY_F5       = 15
	KEY_F6       = 17
	KEY_F7       = 18
	KEY_F8       = 19
	KEY_F9       = 20
	KEY_F10      = 21
	KEY_F11      = 23
	KEY_F12      = 24
)
const (
	// The value of the next keys doesn't matter
	KEY_UP = iota + 24
	KEY_DOWN
	KEY_RIGHT
	KEY_ESC  = ESC // 27
	KEY_LEFT = iota + 24
	KEY_HOME
	KEY_END
)

// /Events

func NewParser(handler event.Eventable) Parser {
	var parser Parser

	parser.handler = handler

	return parser
}

func (parser *Parser) HandleMouseEvent() {
	var ev MouseEvent

	ev.Button = parser.csi_argv[1] & 3
	ev.Mod = parser.csi_argv[1] & 28
	ev.X = parser.csi_argv[2] - 33
	ev.Y = parser.csi_argv[3] - 33

	parser.handler.HandleEvent(ev)
}

func (parser *Parser) HandleComplexCSI() {
	n, err := strconv.Atoi(string(parser.csi_argv[:parser.csi_argc]))

	if err != nil {
		return // Invalid CSI sequence
	}

	if n < KEY_INSERT || n > KEY_F12 {
		return // Unknown key (there are other unknown keys but we won't test them all)
	}

	parser.handler.HandleEvent(SpecialKeyEvent(byte(n)))
}

func CharKeyEvent(ch byte, mod byte) KeyEvent {
	key := ch

	if key >= 'a' && key <= 'z' {
		key -= ' '
	} else if key >= 'A' && key <= 'Z' {
		mod |= MOD_SHIFT
	}

	if key&CTRL == key {
		key |= '@'
		mod |= MOD_CTRL
	}

	return KeyEvent{ch, key, mod}
}

func SpecialKeyEvent(key byte) KeyEvent {
	return KeyEvent{0, key, 0}
}

func (parser *Parser) HandleChar(ch byte) {
	if parser.escape {
		parser.escape = false
		if time.Since(parser.escape_time) < ALT_TIMEOUT {
			if ch == '[' {
				parser.csi = true
				parser.csi_argc = 0
			} else if ch == 'O' {
				parser.ss3 = true
			} else {
				parser.handler.HandleEvent(CharKeyEvent(ch, MOD_ALT))
			}
			return
		}
		parser.handler.HandleEvent(SpecialKeyEvent(ESC))
	} else if parser.csi {
		if parser.csi_argc >= CSI_ARGV_BUFFER_SIZE {
			parser.csi = false
			return
		}

		parser.csi_argv[parser.csi_argc] = ch
		parser.csi_argc++

		if parser.csi_argv[0] == 'M' && parser.csi_argc == 4 {
			parser.HandleMouseEvent()
		} else {
			switch ch {
			case 'A':
				parser.handler.HandleEvent(SpecialKeyEvent(KEY_UP))
			case 'B':
				parser.handler.HandleEvent(SpecialKeyEvent(KEY_DOWN))
			case 'C':
				parser.handler.HandleEvent(SpecialKeyEvent(KEY_RIGHT))
			case 'D':
				parser.handler.HandleEvent(SpecialKeyEvent(KEY_LEFT))
			case 'H':
				parser.handler.HandleEvent(SpecialKeyEvent(KEY_HOME))
			case 'F':
				parser.handler.HandleEvent(SpecialKeyEvent(KEY_END))
			case '~':
				parser.HandleComplexCSI()
			default:
				return
			}
		}

		parser.csi = false
		return
	} else if parser.ss3 {
		parser.ss3 = false

		var key byte

		switch ch {
		case 'P':
			key = KEY_F1
		case 'Q':
			key = KEY_F2
		case 'R':
			key = KEY_F3
		case 'S':
			key = KEY_F4
		default:
			return
		}

		parser.handler.HandleEvent(SpecialKeyEvent(key))
		return
	}

	if ch == ESC {
		parser.escape = true
		parser.escape_time = time.Now()
		return
	}

	parser.handler.HandleEvent(CharKeyEvent(ch, 0))
}

func (parser *Parser) CheckEscape() {
	if parser.escape && time.Since(parser.escape_time) > ALT_TIMEOUT {
		parser.escape = false
		parser.handler.HandleEvent(SpecialKeyEvent(ESC))
	}
}

func (parser *Parser) HandleInputs() {
	for Hasch() {
		parser.HandleChar(Getch())
	}
	parser.CheckEscape()
}
