package pfxlog

import (
	"os"
	"syscall"
)

func init() {
	/*
	 * https://github.com/sirupsen/logrus/issues/496
	 */
	handle := syscall.Handle(os.Stdout.Fd())
	kernel32DLL := syscall.NewLazyDLL("kernel32.dll")
	setConsoleModeProc := kernel32DLL.NewProc("SetConsoleMode")
	setConsoleModeProc.Call(uintptr(handle), 0x0001|0x0002|0x0004)
}
