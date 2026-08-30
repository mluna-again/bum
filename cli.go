package main

import (
	"fmt"
	"os"
	"strings"
)

var addCmdDescription string
var addCmdPaneID string
var addCmdColor string

func runSubcommand(locked bool, args ...string) (int, bool) {
	if len(args) < 1 {
		return 0, false
	}

	cmd := args[0]
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return 0, false
	}

	switch cmd {
	case "add":
		return addItem(locked, args[1:]...), true

	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", cmd)
		return 1, true
	}
}

func addItem(locked bool, args ...string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Missing requirement: title\n")
		return 1
	}
	title := args[0]
	p := Pane{
		TmuxPaneID:  strings.TrimSpace(addCmdPaneID),
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(addCmdDescription),
		Color:       strings.TrimSpace(addCmdColor),
	}

	if p.TmuxPaneID == "" {
		fmt.Fprintf(os.Stderr, "Missing requirement: -pane\n")
		return 1
	}

	if locked {
		panes, err := readCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while reading cache: %q\n", err.Error())
			return 1
		}

		for i, pp := range panes {
			if p.TmuxPaneID == pp.TmuxPaneID {
				panes[i] = p
			}
		}

		err = writeCache(panes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while saving cache: %q\n", err.Error())
			return 1
		}
	} else {
		err := clientRequestNew(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while saving cache: %q\n", err.Error())
			return 1
		}
	}

	return 0
}
