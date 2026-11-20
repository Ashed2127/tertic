package ui

import (
	"os"
	"syscall"
	"unsafe"
)

// RawKey represents a key event detected during recording
type RawKey int

const (
	KeyNone RawKey = iota
	KeyEnter
	KeyEscape
	KeyCtrlC
	KeyStopCommand
)

// ListenKeyForRecording switches the terminal to raw non-canonical mode,
// listens for keypresses (ENTER to stop, ESC to cancel, Ctrl+C to cancel),
// and restores the previous terminal state when done.
func ListenKeyForRecording(stopChan chan<- RawKey, done <-chan struct{}) {
	fd := int(os.Stdin.Fd())
	var oldState syscall.Termios
	ret, _, _ := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
	isTerminal := (int(ret) == 0)

	if isTerminal {
		raw := oldState
		// Disable ICANON (canonical mode) and ECHO (echoing characters)
		// Leave ISIG enabled or handle signal manually
		raw.Lflag &^= (syscall.ICANON | syscall.ECHO)
		// Non-blocking read with 100ms timeout (VMIN=0, VTIME=1)
		raw.Cc[syscall.VMIN] = 0
		raw.Cc[syscall.VTIME] = 1
		_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&raw)), 0, 0, 0)

		defer func() {
			_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
		}()
	}

	buf := make([]byte, 16)
	for {
		select {
		case <-done:
			return
		default:
		}

		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		b := buf[0]
		// ENTER: '\n' (10) or '\r' (13)
		if b == '\n' || b == '\r' {
			stopChan <- KeyEnter
			return
		}

		// Ctrl+C (3)
		if b == 3 {
			stopChan <- KeyCtrlC
			return
		}

		// ESC (27)
		if b == 27 {
			// If it's a standalone ESC without trailing ANSI escape sequence
			if n == 1 {
				stopChan <- KeyEscape
				return
			}
		}
	}
}
