package terminal

import (
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Craftman2868/go-libs/event"
)

const (
	CSI_ARGV_BUFFER_SIZE = 16
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

type ModEvent interface {
	event.Event
	Shift() bool
	Alt() bool
	Ctrl() bool
	Meta() bool
}

type KeyEvent struct {
	Ch  rune
	Key byte
	Mod uint8
}

func (ev KeyEvent) Name() string {
	return "key"
}

func (ev KeyEvent) Shift() bool {
	return ev.Mod&MOD_SHIFT != 0
}

func (ev KeyEvent) Alt() bool {
	return ev.Mod&MOD_ALT != 0
}

func (ev KeyEvent) Ctrl() bool {
	return ev.Mod&MOD_CTRL != 0
}

func (ev KeyEvent) Meta() bool {
	return ev.Mod&MOD_META != 0
}

type MouseEvent struct {
	Button uint8
	Mod    uint8
	X, Y   uint16
}

func (ev MouseEvent) Shift() bool {
	return ev.Mod&MOD_SHIFT != 0
}

func (ev MouseEvent) Alt() bool {
	return ev.Mod&MOD_ALT != 0
}

func (ev MouseEvent) Ctrl() bool {
	return ev.Mod&MOD_CTRL != 0
}

func (ev MouseEvent) Meta() bool {
	return ev.Mod&MOD_META != 0
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
	MOD_ALT
	MOD_CTRL
	MOD_META
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
	KEY_TAB       = '\t'
	KEY_ENTER     = '\n'
	KEY_BACKSPACE = 127
)
const (
	// The value of the next keys (except KEY_ESC) doesn't matter
	KEY_UP = iota + 24
	KEY_DOWN
	KEY_RIGHT
	KEY_ESC  = ESC // 27
	KEY_LEFT = iota + 24
	KEY_BEGIN
	KEY_END
	KEY_HOME
	// Do not add any other key here: the next value is 32 (SPACE)
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
	parser.handler.HandleEvent(CSIErrorEvent{parser.csi_argv[:parser.csi_argc]})
}

func parseNumber(buf []rune) (n int, err error) {
	if len(buf) == 0 {
		return 0, nil
	}

	n, err = strconv.Atoi(string(buf))

	return
}

func parseCSIArgs(args []rune) (res []int, c rune, err error) {
	/*
		The original sequence is analysed like that:
		  sequence ::= CSI <args> <c>
		  args ::= <arg> [';' <args>]  // max CSI_ARGV_BUFFER_SIZE characters
		  arg: a decimal number (zero or more digits)
		  c: any character which is not a digit and not a ';'
	*/

	var n int
	buf := make([]rune, 0, CSI_ARGV_BUFFER_SIZE-1)

	for _, ch := range args[:len(args)-1] {
		if ch == ';' {
			n, err = parseNumber(buf)
			if err != nil {
				return
			}
			res = append(res, n)
			buf = nil
		} else {
			buf = append(buf, ch)
		}
	}

	n, err = parseNumber(buf)
	if err != nil {
		return
	}
	res = append(res, n)

	c = args[len(args)-1]

	return
}

func parseMod(arg int) byte {
	/*
		arg-1	arg	modifiers
		MCAS
		0001	2	shift
		0010	3 	alt
		0011	4 	shift + alt
		0100	5	ctrl
		0101	6	ctrl + shift
		0110	7	ctrl + alt
		0111	8	ctrl + shift + alt
		1000	9	meta
		1001	10	meta + shift
		1010	11	meta + alt
		1011	12	meta + shift + alt
		1100	13	meta + ctrl
		1101	14	meta + ctrl + shift
		1110	15	meta + ctrl + alt
		1111	16	meta + ctrl + shift + alt
	*/
	var mod byte = 0
	arg--

	if arg&1 != 0 {
		mod |= MOD_SHIFT
	}
	if arg&2 != 0 {
		mod |= MOD_ALT
	}
	if arg&4 != 0 {
		mod |= MOD_CTRL
	}
	if arg&8 != 0 {
		mod |= MOD_META
	}

	return mod
}

func (parser *Parser) handleCSI() {
	args, c, err := parseCSIArgs(parser.csi_argv[:parser.csi_argc])

	if err != nil {
		parser.handleCSIError()
		return // Invalid CSI sequence
	}

	var ev KeyEvent

	if len(args) >= 2 {
		ev.Mod = parseMod(args[1])
	} else {
		ev.Mod = 0
	}

	ev.Key = byte(args[0])

	switch c {
	// For the sequences below ('A' to 'H'), args[0] should be 1 or 0 (no argument) but we won't check
	case 'A':
		ev.Key = KEY_UP
	case 'B':
		ev.Key = KEY_DOWN
	case 'C':
		ev.Key = KEY_RIGHT
	case 'D':
		ev.Key = KEY_LEFT
	case 'E':
		ev.Key = KEY_BEGIN
	case 'F':
		ev.Key = KEY_END
	case 'H':
		ev.Key = KEY_HOME

	case '~':
		if args[0] == 27 && len(args) == 3 {
			switch args[2] {
			case 13:
				ev.Key = '\n'
			case 27:
				ev.Key = ESC
			default:
				parser.handleCSIError()
				return
			}
		} else if args[0] < KEY_INSERT || args[0] > KEY_F12 {
			parser.handleCSIError()
			return // unknown key
		}
	case 'u':
		if ev.Key >= 'a' && ev.Key <= 'z' {
			ev.Key -= ' '
		}

	default:
		parser.handleCSIError()
		return // Unknown last char
	}

	parser.handler.HandleEvent(ev)
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

		if key == KEY_ENTER || key == KEY_TAB {
			// do not consider KEY_ENTER as ctrl+J or KEY_TAB as ctrl+i
		} else if key&CTRL == key {
			key |= '@'
			mod |= MOD_CTRL
		}
	}

	return KeyEvent{ch, key, mod}
}

func specialKeyEvent(key byte) KeyEvent {
	return KeyEvent{0, key, 0}
}

func escapeKeyEvent() KeyEvent {
	return KeyEvent{ESC, ESC, 0}
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
		parser.handler.HandleEvent(escapeKeyEvent())
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
		} else if (ch < '0' || ch > '9') && ch != ';' {
			parser.handleCSI()
		} else {
			return
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
		parser.handler.HandleEvent(escapeKeyEvent())
	}
}

func (parser *Parser) HandleInputs() {
	for Hasch() {
		parser.HandleChar(Getch())
	}
	parser.CheckEscape()
}
