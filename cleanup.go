package main

import (
	"fmt"
	"os"

	"github.com/gofrs/flock"
)

func cleanup(l *flock.Flock) {
	err := l.Unlock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", err.Error())
	}
	err = os.Remove(BUM_LOCK)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", err.Error())
	}
	err = os.Remove(BUM_PID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", err.Error())
	}
}
