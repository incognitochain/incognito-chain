// Copyright (c) 2013-2014 The thaibaoautonomous developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// +build !windows,!plan9

package limits

import (
	"fmt"
	"syscall"
)

const (
	FileLimitWant = 2048
	FileLimitMin  = 1024
)

// SetLimits raises some process limits to values which allow node and
// associated utilities to run.
func SetLimits() error {
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}
	if rLimit.Cur > FileLimitWant {
		return nil
	}
	if rLimit.Max < FileLimitMin {
		err = fmt.Errorf("need at least %v file descriptors",
			FileLimitMin)
		return err
	}
	if rLimit.Max < FileLimitWant {
		rLimit.Cur = rLimit.Max
	} else {
		rLimit.Cur = FileLimitWant
	}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		// try min value
		rLimit.Cur = FileLimitMin
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return err
		}
	}

	return nil
}
