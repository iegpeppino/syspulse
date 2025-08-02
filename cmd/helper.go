package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// Helper functions and structs

// Model Initializer
// Creates templates for tables, charts and other models
// to initialize the bubbletea model
func modelInit() model {

	termWidth, termHeight := getTermSize()

	cpuProgress := progress.New(progress.WithGradient("#008000", "#FF0000"), progress.WithoutPercentage())
	cpuProgress.Width = termWidth / 2

	memProgress := progress.New(progress.WithGradient("#008000", "#FF0000"), progress.WithoutPercentage())
	memProgress.Width = termWidth / 2

	cpuColumns := []table.Column{
		{Title: "CPU time", Width: termWidth / 10},
		{Title: "Value (%)", Width: termWidth / 10},
		{Title: "Delta", Width: termWidth / 12},
	}

	cpuTable := initTable(cpuColumns)

	memCols := []table.Column{
		{Title: "Type", Width: termWidth / 8},
		{Title: "Value", Width: termWidth / 8},
	}

	memTable := initTable(memCols)

	procCols := []table.Column{
		{Title: "PID", Width: 5},
		{Title: "Name", Width: 25},
		{Title: "Status", Width: 15},
		{Title: "Runtime", Width: 20},
		{Title: "Memory", Width: 10},
		{Title: "CPU", Width: 10},
	}

	procTable := initTable(procCols)

	diskCols := []table.Column{
		{Title: "Partition", Width: 25},
		{Title: "FsType", Width: 20},
		{Title: "Total", Width: 15},
		{Title: "Used", Width: 15},
		{Title: "Free", Width: 15},
	}

	diskTable := initTable(diskCols)

	m := model{
		tabs:        []string{"CPU", "MEMORY", "PROCESSES", "DISK"},
		ActiveTab:   0,
		keys:        keys,
		help:        help.New(),
		cpuTable:    cpuTable,
		cpuChart:    sparkline.New(termWidth/2, 10, sparkline.WithMaxValue(100.0)),
		cpuProgress: cpuProgress,
		memTable:    memTable,
		memChart:    sparkline.New(termWidth/2, 10, sparkline.WithMaxValue(100.0)),
		memProgress: memProgress,
		procTable:   procTable,
		diskTable:   diskTable,
		width:       termWidth,
		height:      termHeight,
	}

	return m
}

// Create table with default parameters
func initTable(cols []table.Column) table.Model {
	t := table.New(
		table.WithFocused(false),
		table.WithHeight(20),
		table.WithColumns(cols),
		table.WithRows([]table.Row{}),
		table.WithStyles(TableStyle()),
	)
	return t
}

// Returns the terminal width
func getTermSize() (int, int) {
	fd := uintptr(os.Stdout.Fd())
	width, height, _ := term.GetSize(fd)
	return width, height
}

// Compares previous and actual number and returns symbol
// Used to express CPU load variation tendency
func delta(now, prev float64) string {
	d := now - prev
	if d > 0 {
		return "↑"
	} else if d < 0 {
		return "↓"
	} else {
		return "="
	}
}

// Returns the contents of a tab to render
// depending on the activeTab variable
func (m model) renderTab(activeTab int) string {
	switch {
	// CPU stats
	case activeTab == 0:
		return pageContentStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					lipgloss.JoinVertical(
						lipgloss.Center,
						chartTextStyle.Render(lipgloss.PlaceVertical(0, lipgloss.Top, "100%")),
						chartTextStyle.Render(lipgloss.PlaceVertical(10, lipgloss.Bottom, "0%")),
					),
					chartStyle.Render(m.cpuChart.View()),
				),
				gaugeStyle.Render(m.cpuProgress.ViewAs(math.Min(m.cpuTotalPercent/100.0, 1.0))),
				tabGap.Render(strings.Repeat(" ", max(0, m.width))),
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					lipgloss.JoinVertical(
						lipgloss.Top,
						listTitleStyle.Render("CPU Info"),
						listTextStyle.Render(fmt.Sprintf("CPU: %2.f%%\n\nSpeed: %.1f Mhz\n\nCores: %d\n\n%s\n\n",
							m.cpuTotalPercent, m.cpuInfo.Mhz, m.cpuInfo.Cores, m.cpuInfo.ModelName)),
					),
					bottomColStyle.Render(m.cpuTable.View()),
				),
			),
		)
	// Ram stats
	case activeTab == 1:
		return pageContentStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					lipgloss.JoinVertical(
						lipgloss.Center,
						chartTextStyle.Render(lipgloss.PlaceVertical(0, lipgloss.Top, "100%")),
						chartTextStyle.Render(lipgloss.PlaceVertical(10, lipgloss.Bottom, "0%")),
					),
					chartStyle.Render(m.memChart.View()),
				),
				gaugeStyle.Render(m.memProgress.ViewAs(math.Min(m.memory.UsedPercent/100.0, 1.0))),
				tabGap.Render(strings.Repeat(" ", max(0, m.width))),
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					lipgloss.JoinVertical(
						lipgloss.Top,
						listTitleStyle.Render("Memory Info"),
						listTextStyle.Render(fmt.Sprintf("Used RAM: %.2f%%\n", m.memory.UsedPercent)),
					),
					bottomColStyle.Render(m.memTable.View()),
				),
			))
	// Running processes
	case activeTab == 2:
		return pageContentStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("TOP RUNNING PROCESSES"),
			bottomColStyle.Render(m.procTable.View()),
		))
	// Disk availability
	case activeTab == 3:
		return pageContentStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("AVAILABLE DISK PARTITIONS"),
			bottomColStyle.Render(m.diskTable.View()),
		))
	default:
		return fmt.Sprint(m.tabs)
	}
}

// Returns a string expressing bytes along an appropiate magnitude
// otherwise any Memory stat would be a thousand characters long
func getByteMagnitude(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
