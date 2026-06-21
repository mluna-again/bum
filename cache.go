package main

import (
	"encoding/json"
	"os"

	tea "charm.land/bubbletea/v2"
)

type cacheMsg struct {
	panes []Pane
}
func loadCache() tea.Msg {
	data, err := os.ReadFile(BUM_CACHE)
	if err != nil || string(data) == "" {
		return cacheMsg{}
	}

	panes := []Pane{}
	err = json.Unmarshal(data, &panes)
	if err != nil {
		return cacheMsg{}
	}

	return cacheMsg{panes: panes}
}
