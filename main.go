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
	"github.com/mluna-again/luna/luna"
)

type Pane struct {
	TmuxPaneID    string `json:"pane_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	NeedsAtention bool   `json:"needs_attention"`
	Color         string `json:"color"`
}

type model struct {
	panes      []Pane
	termW      int
	termH      int
	selected   int
	errMessage string
	luna       luna.LunaModel
}

func initialModel(l luna.LunaModel) model {
	return model{
		panes:    []Pane{},
		termW:    80,
		termH:    10,
		selected: -1,
		luna:     l,
	}
}

func (m model) Init() tea.Cmd {
	return m.luna.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case serverNewPaneMsg:
		index := -1
		for i, p := range m.panes {
			if p.TmuxPaneID == msg.pane.TmuxPaneID {
				index = i
			}
		}
		if index == -1 {
			m.panes = append(m.panes, msg.pane)
		} else {
			m.panes[index].Title = msg.pane.Title
			m.panes[index].Description = msg.pane.Description
			m.panes[index].NeedsAtention = msg.pane.NeedsAtention
			m.panes[index].Color = msg.pane.Color
		}
		return m, nil

	case focusPaneMsg:
		m.selected = -1
		if msg.err != nil {
			m.errMessage = msg.err.Error()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.termH = msg.Height
		m.termW = msg.Width
		return m, nil

	case tea.BlurMsg:
		m.selected = -1
		return m, nil

	case tea.FocusMsg:
		if m.selected == -1 {
			break
		}
		m.errMessage = ""
		return m, focusPane(m.panes[m.selected])

	case tea.MouseReleaseMsg:
		if !zone.Get(fmt.Sprintf("%d", m.selected)).InBounds(msg) {
			break
		}
		if m.selected == -1 {
			break
		}
		m.errMessage = ""
		return m, focusPane(m.panes[m.selected])

	case tea.MouseMotionMsg:
		newFocused := false

		// prevents "ghost" switches (on the tea.FocusMsg event) when the mouse goes out of the window
		// leaving a row selected
		if msg.X >= m.termW-2 || msg.X < 1 {
			m.selected = -1
			return m, nil
		}
		for i := range m.panes {
			id := fmt.Sprintf("%d", i)
			info := zone.Get(id)
			if info.InBounds(msg) {
				m.selected = i
				newFocused = true
				break
			}
		}
		if !newFocused {
			m.selected = -1
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
			if m.selected == -1 {
				break
			}
			m.errMessage = ""
			return m, focusPane(m.panes[m.selected])

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.luna, cmd = m.luna.Update(msg)

	return m, cmd
}

func (m model) View() tea.View {
	elements := []string{}
	if len(m.panes) == 0 {
		elements = append(elements, lipgloss.PlaceHorizontal(m.termW, lipgloss.Center, "No panes tagged yet!"))
	}
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

		elements = append(elements, zone.Mark(fmt.Sprintf("%d", i), c))
	}

	l := m.luna.View().Content
	l = lipgloss.PlaceHorizontal(m.termW, lipgloss.Center, l)
	lh := lipgloss.Height(l)

	content := lipgloss.JoinVertical(lipgloss.Top, elements...)
	content = lipgloss.PlaceVertical(m.termH-lh-1, lipgloss.Top, content)
	content = lipgloss.JoinVertical(lipgloss.Top, content, l, m.errMessage)

	return tea.View{
		Content:     zone.Scan(content),
		AltScreen:   true,
		MouseMode:   tea.MouseModeAllMotion,
		ReportFocus: true,
	}
}

func main() {
	l, errs := luna.NewLuna(luna.NewLunaParams{
		Animation: luna.LunaAnimation("sleeping"),
		Pet:       luna.LunaPet("cat"),
		Variant:   luna.LunaVariant("ragdoll"),
		Size:      luna.SMALL,
	})
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Error initializing luna:")
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		return
	}

	zone.NewGlobal()
	p := tea.NewProgram(initialModel(l))
	go startServer(p)
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
