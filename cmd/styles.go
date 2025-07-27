package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func TableStyle() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		Margin(2, 0, 0, 0).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FFBF00")).
		BorderBottom(true).
		AlignVertical(lipgloss.Center).
		Bold(false)
	s.Cell = s.Cell.
		AlignHorizontal(lipgloss.Left).
		Padding(1, 0, 0, 1)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Bold(false)
	return s
}

// Defining used colors
var (
	normal = lipgloss.Color("#EEEEEE")
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	green  = lipgloss.Color("#139213ff")
	yellow = lipgloss.Color("#f1f155ff")
	orange = lipgloss.Color("#FFA500")
	red    = lipgloss.Color("#d62222ff")
	gray   = lipgloss.Color("#444444")
	white  = lipgloss.Color("#FBFBFB")
	amber  = lipgloss.Color("#FFBF00")
)

var (
	chartStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(amber)).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(amber))

	chartTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(normal)).
			Align(lipgloss.Right)

	gaugeGapStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(amber)).
			BorderBottom(true).
			BorderRight(true).
			MarginRight(1)

	gaugeBarStyle = gaugeGapStyle.
			BorderBottom(false).
			BorderRight(false).
			MarginRight(0)

	gaugeTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(normal))

	baseStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#FFBF00")).
			Bold(true).
			Padding(1, 1, 1, 2).
			Margin(0, 0, 0, 2).
			AlignHorizontal(lipgloss.Center)

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	base = lipgloss.NewStyle().Foreground(normal)

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(amber).
		Padding(0, 1)

	activeTab = tab.Border(activeTabBorder, true)

	tabGap = tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	gauge = lipgloss.NewStyle().
		Foreground(normal).
		Margin(1, 1).
		Padding(1, 1)

	titleStyle = lipgloss.NewStyle().
			Margin(2, 5, 1, 5).
			Padding(0, 1, 0, 1).
			Italic(true).
			Bold(true).
			Foreground(lipgloss.Color(normal)).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(amber).
			BorderBottom(true)

	pageContentStyle = lipgloss.NewStyle().
				Height(32)
)

//pageContentStyle.Render()

func gaugeProgress(cpuPercent float64) lipgloss.Color {
	switch {
	case cpuPercent < 50:
		return green
	case cpuPercent < 75:
		return yellow
	case cpuPercent < 90:
		return orange
	default:
		return red
	}
}
