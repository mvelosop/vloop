//go:build !windows

package driver

import "errors"

// osHold is the Windows hold; there is none here.
func osHold() error { return errors.New("not Windows") }
