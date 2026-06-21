package main

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

type Pane struct {
	TmuxPaneID     string
	Title string
	Description    string
	NeedsAtention  bool
	Color          string
}

type model struct {
	panes      []Pane
	termW      int
	termH      int
	selected   int
	focused    bool
	errMessage string
}

func initialModel() model {
	return model{
		panes: []Pane{
			{"config:1", "tmux", "config files", false, "2"},
			{"config:5", "nvim", "config files", false, "2"},
			{"bum:1", "bum", "source code", false, "2"},
			{"storage:4", "yazi", "zfs dataset", false, "3"},
			{"void:1", "the void", "idle", false, "1"},
		},
		termW:    80,
		termH:    10,
		selected: 0,
		focused:  true,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case focusPaneMsg:
		if msg.err != nil {
			m.errMessage = msg.err.Error()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.termH = msg.Height
		m.termW = msg.Width
		return m, nil

	case tea.FocusMsg:
		m.focused = true
		return m, nil

	case tea.BlurMsg:
		m.focused = false
		return m, nil

	case tea.MouseReleaseMsg:
		return m, focusPane(m.panes[m.selected])

	case tea.MouseMsg:
		for i := range m.panes {
			if zone.Get(fmt.Sprintf("%d", i)).InBounds(msg) {
				m.selected = i
				break
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			m.selected++
			if m.selected > len(m.panes)-1 {
				m.selected = 0
			}
			return m, nil

		case "k", "up":
			m.selected--
			if m.selected < 0 {
				m.selected = len(m.panes) - 1
			}
			return m, nil

		case "enter":
			return m, focusPane(m.panes[m.selected])

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	cards := []string{}
	for i, p := range m.panes {
		cs := Card
		if i == m.selected {
			cs = HoveredCard
		}
		indicator := Indicator.Inherit(cs).Render(" ")
		if p.Color != "" {
			indicator = Indicator.Foreground(lipgloss.Color(p.Color)).Inherit(cs).Render("●")
		}
		title := Title.Inherit(cs).Render(p.Title)
		description := Description.Inherit(cs).Render(p.Description)
		firstLine := lipgloss.JoinHorizontal(lipgloss.Left, indicator, title)
		firstLine = lipgloss.PlaceHorizontal(m.termW, lipgloss.Left, firstLine, lipgloss.WithWhitespaceStyle(cs))
		secondLine := lipgloss.PlaceHorizontal(m.termW, lipgloss.Left, description, lipgloss.WithWhitespaceStyle(cs))
		c := lipgloss.JoinVertical(lipgloss.Top, firstLine, secondLine)

		cards = append(cards, zone.Mark(fmt.Sprintf("%d", i), c))
	}

	content := lipgloss.JoinVertical(lipgloss.Top, cards...)
	content = lipgloss.PlaceVertical(m.termH-1, lipgloss.Top, content)
	content = lipgloss.JoinVertical(lipgloss.Top, content, m.errMessage)
	return tea.View{
		Content:   zone.Scan(content),
		AltScreen: true,
		MouseMode: tea.MouseModeAllMotion,
	}
}

func main() {
	zone.NewGlobal()
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type focusPaneMsg struct {
	err error
}

func focusPane(p Pane) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("tmux", "switch-client", "-t", p.TmuxPaneID)
		var out bytes.Buffer
		var serr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &serr
		err := cmd.Run()
		var prettyErr error
		if err != nil {
			prettyErr = errors.New(cmp.Or(serr.String(), err.Error()))
		}

		return focusPaneMsg{
			err: prettyErr,
		}
	}
}
