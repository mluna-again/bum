package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

type Pane struct {
	TmuxPaneID     string
	RunningProgram string
	Description    string
	NeedsAtention  bool
	Color          string
}

type model struct {
	panes    []Pane
	termW    int
	termH    int
	selected int
}

func initialModel() model {
	return model{
		panes: []Pane{
			{"%1", "nvim", "", false, ""},
			{"%1", "nvim", "idk", false, ""},
			{"%1", "tail", "watching logs", false, "3"},
			{"%1", "grep", "searching logs", true, "2"},
			{"%1", "btm", "idle", false, "1"},
		},
		termW:    80,
		termH:    10,
		selected: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termH = msg.Height
		m.termW = msg.Width
		return m, nil
	
	case tea.MouseMsg:
		for i := range m.panes {
			if zone.Get(fmt.Sprintf("%d", i)).InBounds(msg) {
				m.selected = i
				return m, nil
			}
		}

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
		title := Title.Inherit(cs).Render(p.RunningProgram)
		description := Description.Inherit(cs).Render(p.Description)
		firstLine := lipgloss.JoinHorizontal(lipgloss.Left, indicator, title)
		firstLine = lipgloss.PlaceHorizontal(m.termW, lipgloss.Left, firstLine, lipgloss.WithWhitespaceStyle(cs))
		secondLine := lipgloss.PlaceHorizontal(m.termW, lipgloss.Left, description, lipgloss.WithWhitespaceStyle(cs))
		c := lipgloss.JoinVertical(lipgloss.Top, firstLine, secondLine)

		cards = append(cards, zone.Mark(fmt.Sprintf("%d", i), c))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, cards...)
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
