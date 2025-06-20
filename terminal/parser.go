package terminal

import (
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Craftman2868/go-libs/event"
)

const (
	CSI_ARGV_BUFFER_SIZE = 8
	ALT_TIMEOUT          = 50 * time.Millisecond
)

const ESC = 27

type Parser struct {
	handler event.Eventable

	buf []byte

	escape      bool
	escape_time time.Time

	csi      bool
	csi_argv [CSI_ARGV_BUFFER_SIZE]rune
	csi_argc uint8

	ss3 bool
}

// Events

type KeyEvent struct {
	Ch  rune
	Key byte
	Mod uint8
}

func (ev KeyEvent) Name() string {
	return "key"
}

func (ev KeyEvent) Ctrl() bool {
	return ev.Mod&MOD_CTRL != 0
}

func (ev KeyEvent) Alt() bool {
	return ev.Mod&MOD_ALT != 0
}

func (ev KeyEvent) Shift() bool {
	return ev.Mod&MOD_SHIFT != 0
}

type MouseEvent struct {
	Button uint8
	Mod    uint8
	X, Y   uint16
}

func (ev MouseEvent) Name() string {
	return "mouse"
}

type UnicodeErrorEvent struct {
	Buf []byte
}

func (ev UnicodeErrorEvent) Name() string {
	return "unicodeError"
}

type CSIErrorEvent struct {
	Buf []rune
}

func (ev CSIErrorEvent) Name() string {
	return "CSIError"
}

const ( // MouseEvent.Button
	BUTTON1 = iota
	BUTTON2
	BUTTON3
	RELEASE
	SCROLL_UP
	SCROLL_DOWN
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

func (parser *Parser) handleMouseEvent() {
	var ev MouseEvent
	ev.Button = uint8(int32(parser.csi_argv[1]) & 3)
	ev.Mod = uint8(int32(parser.csi_argv[1]) & 28)
	ev.X = uint16(int32(parser.csi_argv[2]) - 33)
	ev.Y = uint16(int32(parser.csi_argv[3]) - 33)

	if int32(parser.csi_argv[1])&64 != 0 {
		ev.Button += 4
	}

	parser.handler.HandleEvent(ev)
}

func (parser *Parser) handleCSIError() {
	parser.handler.HandleEvent(CSIErrorEvent{parser.csi_argv[:]})
}

func parseCSIArgs(args []rune) (res []int, c rune, err error) {
	var buf []rune
	var n int

	c = args[len(args)-1]

	for _, ch := range args[:len(args)-1] {
		if ch == ';' {
			n, err = strconv.Atoi(string(buf))
			if err != nil {
				return
			}
			res = append(res, n)
			buf = nil
		} else {
			buf = append(buf, ch)
		}
	}

	n, err = strconv.Atoi(string(buf))
	if err != nil {
		return
	}
	res = append(res, n)

	return
}

func (parser *Parser) handleComplexCSI() {
	args, c, err := parseCSIArgs(parser.csi_argv[:parser.csi_argc])

	if err != nil {
		parser.handleCSIError()
		return // Invalid CSI sequence
	}

	switch c {
	case '~':
		if len(args) != 1 {
			parser.handleCSIError()
			return // Unknown sequence
		}
		if args[0] < KEY_INSERT || args[0] > KEY_F12 {
			parser.handleCSIError()
			return // Unknown key (there are other unknown keys but we won't test them all)
		}
		parser.handler.HandleEvent(specialKeyEvent(byte(args[0])))
	case 'u':
		if len(args) != 2 {
			parser.handleCSIError()
			return // Unknown sequence
		}
		var ev KeyEvent

		ev.Mod = MOD_CTRL

		if (args[1]-5)&1 != 0 {
			ev.Mod |= MOD_SHIFT
		}
		if (args[1]-5)&2 != 0 {
			ev.Mod |= MOD_ALT
		}
		/*
			00 5 -> ctrl
			01 6 -> ctrl + shift
			10 7 -> ctrl + alt
			11 8 -> ctrl + alt + shift
		*/

		ev.Ch = rune(args[0])
		ev.Key = byte(args[0])
		if ev.Key >= 'a' && ev.Key <= 'z' {
			ev.Key -= ' '
		}
		parser.handler.HandleEvent(ev)
	default:
		parser.handleCSIError()
		return // Unknown last char
	}
}

func charKeyEvent(ch rune, mod byte) KeyEvent {
	var key byte

	if ch < unicode.MaxLatin1 {
		key = byte(ch)

		if key >= 'a' && key <= 'z' {
			key -= ' '
		} else if key >= 'A' && key <= 'Z' {
			mod |= MOD_SHIFT
		}

		if key&CTRL == key {
			key |= '@'
			mod |= MOD_CTRL
		}
	}

	return KeyEvent{ch, key, mod}
}

func specialKeyEvent(key byte) KeyEvent {
	return KeyEvent{0, key, 0}
}

func (parser *Parser) HandleRune(ch rune) {
	if parser.escape {
		parser.escape = false
		if time.Since(parser.escape_time) < ALT_TIMEOUT {
			if ch == '[' {
				parser.csi = true
				parser.csi_argc = 0
			} else if ch == 'O' {
				parser.ss3 = true
			} else {
				parser.handler.HandleEvent(charKeyEvent(ch, MOD_ALT))
			}
			return
		}
		parser.handler.HandleEvent(specialKeyEvent(ESC))
	} else if parser.csi {
		if parser.csi_argc >= CSI_ARGV_BUFFER_SIZE {
			parser.csi = false
			parser.handler.HandleEvent(CSIErrorEvent{parser.csi_argv[:parser.csi_argc]})
			parser.HandleRune(ch)
			return
		}

		parser.csi_argv[parser.csi_argc] = ch
		parser.csi_argc++

		if parser.csi_argv[0] == 'M' {
			if parser.csi_argc == 4 {
				parser.handleMouseEvent()
			} else {
				return
			}
		} else {
			switch ch {
			case 'A':
				parser.handler.HandleEvent(specialKeyEvent(KEY_UP))
			case 'B':
				parser.handler.HandleEvent(specialKeyEvent(KEY_DOWN))
			case 'C':
				parser.handler.HandleEvent(specialKeyEvent(KEY_RIGHT))
			case 'D':
				parser.handler.HandleEvent(specialKeyEvent(KEY_LEFT))
			case 'H':
				parser.handler.HandleEvent(specialKeyEvent(KEY_HOME))
			case 'F':
				parser.handler.HandleEvent(specialKeyEvent(KEY_END))
			case '~':
				fallthrough
			case 'u':
				parser.handleComplexCSI()
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
			return // unknown ss3 sequence
		}

		parser.handler.HandleEvent(specialKeyEvent(key))
		return
	}

	if ch == ESC {
		parser.escape = true
		parser.escape_time = time.Now()
		return
	}

	parser.handler.HandleEvent(charKeyEvent(ch, 0))
}

func (parser *Parser) HandleChar(ch byte) {
	// fmt.Printf("char: %d", ch)
	// if ch >= 32 {
	// 	fmt.Printf(", '%c'", ch)
	// }
	// fmt.Println()

	parser.buf = append(parser.buf, ch)

	if !utf8.FullRune(parser.buf) {
		return
	}

	r, size := utf8.DecodeRune(parser.buf)

	if r == utf8.RuneError {
		parser.handler.HandleEvent(UnicodeErrorEvent{parser.buf})
		parser.buf = parser.buf[:0]
		return
	}

	parser.buf = parser.buf[size:]

	parser.HandleRune(r)
}

func (parser *Parser) CheckEscape() {
	if parser.escape && time.Since(parser.escape_time) > ALT_TIMEOUT {
		parser.escape = false
		parser.handler.HandleEvent(specialKeyEvent(ESC))
	}
}

func (parser *Parser) HandleInputs() {
	for Hasch() {
		parser.HandleChar(Getch())
	}
	parser.CheckEscape()
}
