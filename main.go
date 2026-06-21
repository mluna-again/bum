package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/gofrs/flock"
	zone "github.com/lrstanley/bubblezone/v2"
	"github.com/mluna-again/luna/luna"
)

var toggle bool

// there is probably a better way of doing this but whatever
const BUM_LOCK = "/tmp/bum-4f766dad-c62f-4102-9f0e-87c27d054f35.lock"
const BUM_PID = "/tmp/bum-4f766dad-c62f-4102-9f0e-87c27d054f35.pid"
const BUM_CACHE = "/tmp/bum-4f766dad-c62f-4102-9f0e-87c27d054f35.cache"

type Pane struct {
	TmuxPaneID    string `json:"pane_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	NeedsAtention bool   `json:"needs_attention"`
	Color         string `json:"color"`
}

type model struct {
	panes       []Pane
	termW       int
	termH       int
	selected    int
	deleteHover bool
	errMessage  string
	luna        luna.LunaModel
	ready       bool
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
	return tea.Batch(m.luna.Init(), loadCache)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cacheMsg:
		m.panes = msg.panes
		m.ready = true
		return m, nil

	case serverNewPaneMsg:
		if !m.ready {
			break
		}
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
		f, err := os.Create(BUM_CACHE)
		if err != nil {
			m.errMessage = err.Error()
			return m, nil
		}
		defer f.Close()
		e := json.NewEncoder(f)
		err = e.Encode(m.panes)
		if err != nil {
			m.errMessage = err.Error()
			return m, nil
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
		if zone.Get(fmt.Sprintf("%d-delete", m.selected)).InBounds(msg) {
			newPanes := []Pane{}
			for i, pane := range m.panes {
				if i == m.selected {
					continue
				}
				newPanes = append(newPanes, pane)
			}
			m.panes = newPanes
			m.selected = -1
			m.deleteHover = false
			return m, nil
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
				if zone.Get(fmt.Sprintf("%d-delete", i)).InBounds(msg) {
					m.deleteHover = true
				} else {
					m.deleteHover = false
				}
				newFocused = true
				break
			}
		}
		if !newFocused {
			m.selected = -1
			m.deleteHover = false
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
		ds := Card
		if i == m.selected {
			cs = HoveredCard
			ds = HoveredCard
		}
		if i == m.selected && m.deleteHover {
			ds = InnerHoveredCard
		}
		indicator := Indicator.Inherit(cs).Render(" ")
		if p.Color != "" {
			indicator = Indicator.Foreground(lipgloss.Color(p.Color)).Inherit(cs).Render("●")
		}
		title := Title.Inherit(cs).Render(p.Title)
		description := Description.Inherit(cs).Render(p.Description)
		deleteIcon := ""
		if i == m.selected {
			deleteIcon = Delete.Inherit(ds).Render(" Delete ")
		}
		deleteIcon = zone.Mark(fmt.Sprintf("%d-delete", i), deleteIcon)
		firstLine := lipgloss.JoinHorizontal(lipgloss.Left, indicator, title)
		firstLine = lipgloss.PlaceHorizontal(m.termW-lipgloss.Width(deleteIcon), lipgloss.Left, firstLine, lipgloss.WithWhitespaceStyle(cs))
		firstLine = lipgloss.JoinHorizontal(lipgloss.Left, firstLine, deleteIcon)
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
	flag.BoolVar(&toggle, "toggle", false, "start bum or kill current running instance")
	flag.Parse()

	lock := flock.New(BUM_LOCK)
	locked, err := lock.TryLock()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	if !locked {
		if !toggle {
			fmt.Fprintln(os.Stderr, "another instance of bum is already running")
			os.Exit(1)
		}
		data, err := os.ReadFile(BUM_PID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		pid, err := strconv.Atoi(string(data))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
		err = proc.Signal(os.Interrupt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while killing other bum instance: %s", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
	defer cleanup(lock)

	pid, err := os.Create(BUM_PID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	_, err = pid.WriteString(fmt.Sprintf("%d", os.Getpid()))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	err = pid.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

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
		cleanup(lock)
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
