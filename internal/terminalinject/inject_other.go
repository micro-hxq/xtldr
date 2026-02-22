//go:build !darwin && !linux && !freebsd && !netbsd && !openbsd && !dragonfly

package terminalinject

import "errors"

func InsertIntoCommandLine(command string) error {
	_ = command
	return errors.New("terminal command line insertion is not supported on this platform")
}
