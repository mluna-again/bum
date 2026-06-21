package main

import "charm.land/lipgloss/v2"

var BG = lipgloss.Color("8")
var Indicator = lipgloss.NewStyle().Padding(0, 1)
var Title = lipgloss.NewStyle().Foreground(lipgloss.White).PaddingRight(1)
var Description = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(3)
var Card = lipgloss.NewStyle()
var HoveredCard = lipgloss.NewStyle().Background(BG)
