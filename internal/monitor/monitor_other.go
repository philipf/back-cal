//go:build !windows

package monitor

import "errors"

func PrimaryMonitor() (width, height int, err error) {
	return 0, 0, errors.New("unsupported OS")
}
