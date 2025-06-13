package terminal

import (
	"os"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

const STDIN = 0
const CTRL = 31

var old_termios unix.Termios

func InitInput() {
	err := termios.Tcgetattr(STDIN, &old_termios)

	if err != nil {
		panic("Error in Tcgetattr: " + err.Error())
	}

	cur_termios := old_termios

	cur_termios.Lflag &= ^uint32(unix.ICANON | unix.ECHO | unix.ISIG)
	cur_termios.Iflag &= ^uint32(unix.IXON)

	err = termios.Tcsetattr(STDIN, termios.TCSAFLUSH, &cur_termios)

	if err != nil {
		panic("Error in Tcsetattr: " + err.Error())
	}
}

func QuitInput() {
	err := termios.Tcsetattr(STDIN, termios.TCSAFLUSH, &old_termios)

	if err != nil {
		panic("Error in Tcsetattr: " + err.Error())
	}
}

func Getch() byte {
	var buf [1]byte

	_, err := os.Stdin.Read(buf[:])

	if err != nil {
		panic("Error in Stdin.Read: " + err.Error())
	}

	return buf[0]
}

func Available() int {
	n, err := unix.IoctlGetInt(STDIN, unix.TIOCINQ)

	if err != nil {
		panic("Error in ioctl(0, TIOCINQ): " + err.Error())
	}

	return n
}

func Hasch() bool {
	return Available() > 0
}

func GetchIfAny() (byte, bool) {
	if Hasch() {
		return Getch(), true
	}

	return 0, false
}

func GetSize() (uint16, uint16) {
	ws, err := unix.IoctlGetWinsize(STDIN, unix.TIOCGWINSZ)

	if err != nil {
		panic("Error in ioctl(0, TIOCGWINSZ): " + err.Error())
	}

	return ws.Col, ws.Row
}
