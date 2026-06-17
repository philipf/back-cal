//go:build !windows

package wallpaper

import "errors"

func Set(path string) error {
	return errors.New("unsupported OS")
}
