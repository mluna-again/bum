package main

import "charm.land/lipgloss/v2"

var BGDARK = lipgloss.Color("0")
var BG = lipgloss.Color("8")
var FG = lipgloss.BrightWhite
var Indicator = lipgloss.NewStyle().Padding(0, 1)
var Title = lipgloss.NewStyle().Foreground(lipgloss.White).PaddingRight(1)
var Description = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(3)
var Delete = lipgloss.NewStyle().PaddingRight(3)
var Card = lipgloss.NewStyle()
var HoveredCard = lipgloss.NewStyle().Background(BG)
var InnerHoveredCard = lipgloss.NewStyle().Background(BGDARK)
var Bordered = lipgloss.NewStyle().Border(lipgloss.ASCIIBorder())
var TitleBar = lipgloss.NewStyle().Background(BGDARK).Foreground(FG)
