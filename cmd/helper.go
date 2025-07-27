package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Helper functions and structs

// Model Initializer
func modelInit() model {

	cpuProgress := progress.New(progress.WithGradient("#008000", "#FF0000"), progress.WithoutPercentage())
	cpuProgress.Width = 40

	cpuColumns := []table.Column{
		{Title: "Load", Width: 15},
		{Title: "Value (%)", Width: 15},
		{Title: "Delta", Width: 10},
	}

	cpuTable := initTable(cpuColumns)

	memCols := []table.Column{
		{Title: "Type", Width: 40},
		{Title: "Value", Width: 40},
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
		cpuProgress: cpuProgress,
		memTable:    memTable,
		procTable:   procTable,
		diskTable:   diskTable,
		cpuChart:    sparkline.New(40, 10, sparkline.WithMaxValue(100.0)),
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

// CPU and Memory load gauge bar constructor
func loadGauge(loadPercent float64, width int) string {
	// Calculate width for current load
	full := int(loadPercent / 100 * float64(width))
	empty := width - full

	bar := strings.Builder{}
	// Render left separator
	bar.WriteString(
		lipgloss.NewStyle().
			Foreground(amber).
			Render(" | "))
	// Render load progress
	for i := 0; i < full; i++ {
		bar.WriteString(
			lipgloss.NewStyle().
				Foreground(gaugeProgress(loadPercent)). // Gets load based color
				Render("█ "))
	}
	// Render the empty part of the gauge bar
	for i := 0; i < empty; i++ {
		bar.WriteString(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(gray)).
				Render("░ "))
	}
	// Render right separator
	bar.WriteString(
		lipgloss.NewStyle().
			Foreground(amber).
			Render(" | "))
	return bar.String()
}

// Returns the contents of a tab to render
// depending on the activeTab variable
func (m model) renderTab(activeTab int) string {
	switch {
	// CPU stats
	case activeTab == 0:
		return pageContentStyle.Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left,
				chartStyle.Render(m.cpuChart.View()),
				m.cpuProgress.ViewAs(math.Min(m.cpuTotalPercent/100.0, 1.0)),
			),
			baseStyle.Render(m.cpuTable.View()),
		),
		)

		// return pageContentStyle.Render(lipgloss.JoinVertical(
		// 	lipgloss.Left,
		// 	gauge.Render(fmt.Sprintf(
		// 		"CPU: %.2f%%\n%s\n",
		// 		m.cpuTotalPercent,
		// 		loadGauge(m.cpuTotalPercent, 45))),
		// 	baseStyle.Render(m.cpuTable.View()),
		// ))
		// return lipgloss.JoinVertical(
		// 	lipgloss.Left,
		// 	gauge.Render(fmt.Sprintf(
		// 		"CPU: %.2f%%\n%s\n",
		// 		m.cpuTotalPercent,
		// 		loadGauge(m.cpuTotalPercent, 45))),
		// 	baseStyle.Render(m.cpuTable.View()),
		// )
	// Ram stats
	case activeTab == 1:
		return pageContentStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			gauge.Render(fmt.Sprintf(
				"RAM: %.2f%%\n%s\n",
				m.memory.UsedPercent,
				loadGauge(m.memory.UsedPercent, 45))),
			baseStyle.Render(m.memTable.View()),
		))
		// return lipgloss.JoinVertical(
		// 	lipgloss.Left,
		// 	gauge.Render(fmt.Sprintf(
		// 		"RAM: %.2f%%\n%s\n",
		// 		m.memory.UsedPercent,
		// 		loadGauge(m.memory.UsedPercent, 45))),
		// 	baseStyle.Render(m.memTable.View()),
		// )
	// Running processes
	case activeTab == 2:
		return pageContentStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("TOP RUNNING PROCESSES"),
			baseStyle.Render(m.procTable.View()),
		))
		// return lipgloss.JoinVertical(
		// 	lipgloss.Left,
		// 	titleStyle.Render("TOP RUNNING PROCESSES"),
		// 	baseStyle.Render(m.procTable.View()),
		// )
	// Disk availability
	case activeTab == 3:
		return pageContentStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("AVAILABLE DISK PARTITIONS"),
			baseStyle.Render(m.diskTable.View()),
		))
		// return lipgloss.JoinVertical(
		// 	lipgloss.Left,
		// 	titleStyle.Render("AVAILABLE DISK PARTITIONS"),
		// 	baseStyle.Render(m.diskTable.View()),
		// )
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
