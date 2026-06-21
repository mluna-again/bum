package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

func (m model) sessionList() string {
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

	return lipgloss.JoinVertical(lipgloss.Top, elements...)
}
