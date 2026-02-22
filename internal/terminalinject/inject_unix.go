//go:build darwin || linux || freebsd || netbsd || openbsd || dragonfly

package terminalinject

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

func InsertIntoCommandLine(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	tiocstiErr := injectWithTIOCSTI(command)
	if tiocstiErr == nil {
		return nil
	}

	if runtime.GOOS == "darwin" {
		if err := injectWithAppleScript(command); err == nil {
			return nil
		} else {
			return fmt.Errorf("TIOCSTI injection failed (%v) and AppleScript injection failed (%w)", tiocstiErr, err)
		}
	}

	return fmt.Errorf("TIOCSTI injection failed: %w", tiocstiErr)
}

func injectWithTIOCSTI(command string) error {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer tty.Close()

	fd := int(tty.Fd())
	for _, b := range []byte(command) {
		if err := unix.IoctlSetPointerInt(fd, unix.TIOCSTI, int(b)); err != nil {
			return err
		}
	}
	return nil
}

func injectWithAppleScript(command string) error {
	script := `on run argv
set cmdText to item 1 of argv
tell application "System Events"
  keystroke cmdText
end tell
end run
`

	cmd := exec.Command("osascript", "-", command)
	cmd.Stdin = strings.NewReader(script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
