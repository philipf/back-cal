//go:build !windows

package scheduler

import "errors"

func Register(executablePath string) error {
	return errors.New("unsupported OS")
}
