package main

import (
	"fmt"
	"github/iegpeppino/syspulse/logger"
	"github/iegpeppino/syspulse/systeminfo"
	"log/slog"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type model struct {
	tabs            []string
	ActiveTab       int
	height          int
	width           int
	keys            keyMap
	help            help.Model
	cpuTotalPercent float64
	cpuInfo         cpu.InfoStat
	cpuStats        cpu.TimesStat
	cpuPrevStats    cpu.TimesStat
	cpuTable        table.Model
	cpuChart        sparkline.Model
	cpuProgress     progress.Model
	processes       []systeminfo.ProcessInfo
	procTable       table.Model
	memory          mem.VirtualMemoryStat
	memTable        table.Model
	memChart        sparkline.Model
	memProgress     progress.Model
	disk            []systeminfo.DiskInfo
	diskTable       table.Model
	err             error
}

// Setup for key bindings
type keyMap struct {
	Left  key.Binding
	Right key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// Setting help message formats
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Help, k.Quit},
	}
}

// Set key bindings using lipgloss.help
var keys = keyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "a"),
		key.WithHelp("←/a", "switch tab to left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "d"),
		key.WithHelp("→/d", "switch tab to right"),
	),
	Help: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type tickMsg struct{}

// Setting the ticker for 500 milliseconds intervals
func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update model height and width
		m.width, m.height = getTermSize()
		m.help.Width = msg.Width

	// In case of tick
	case tickMsg:
		// Get general cpu information
		cpuInfo, err := systeminfo.GetCPUinfo()
		if err != nil {
			logger.Logger.Error("Couldn't get CPU Info", slog.String("error", err.Error()))
		}
		m.cpuInfo = cpuInfo

		//Get and update system stats
		cpuPercent, err := systeminfo.GetCPUPercent()
		if err != nil {
			logger.Logger.Error("Couldn't get CPU times", slog.String("error", err.Error()))
		}

		m.cpuTotalPercent = cpuPercent
		// m.cpuTotalPercent = rand.Float64() * 100 // Only used for visual testing purposes

		// Update and draw CPU usage chart
		m.cpuChart.Push(m.cpuTotalPercent)
		m.cpuChart.Draw()

		mem, err := systeminfo.GetMEMLoad()
		if err != nil {
			logger.Logger.Error("Couldn't get memory stats", slog.String("error", err.Error()))
		}
		m.memory = *mem

		// Update and draw RAM usage chart
		m.memChart.Push(m.memory.UsedPercent)
		m.memChart.Draw()

		// Compare previous cpuTimes with current
		// and observe increment or decrement tendency
		cpuTimes, _ := systeminfo.GetCPUTimes()
		if len(*cpuTimes) > 0 {
			m.cpuPrevStats = m.cpuStats
			prevTimes := (*cpuTimes)[0]
			m.cpuStats = prevTimes
		}

		processes, err := systeminfo.GetProcessInfo(9)
		if err != nil {
			logger.Logger.Error("Unable to read running processes", slog.String("error", err.Error()))
		}
		m.processes = processes

		disks, err := systeminfo.GetDISKUse()
		if err != nil {
			logger.Logger.Error("Disk info error", slog.String("error", err.Error()))
		}
		m.disk = disks

		// Update CPU table information
		cpuRows := []table.Row{
			{"User", fmt.Sprintf("%.2f%%", m.cpuStats.User), delta(m.cpuStats.User, m.cpuPrevStats.User)},
			{"System", fmt.Sprintf("%.2f%%", m.cpuStats.System), delta(m.cpuStats.System, m.cpuPrevStats.System)},
			{"Idle", fmt.Sprintf("%.2f%%", m.cpuStats.Idle), delta(m.cpuStats.Idle, m.cpuPrevStats.Idle)},
			{"Nice", fmt.Sprintf("%.2f%%", m.cpuStats.Nice), delta(m.cpuStats.Nice, m.cpuPrevStats.Nice)},
			{"Guest", fmt.Sprintf("%.2f%%", m.cpuStats.Guest), delta(m.cpuStats.Guest, m.cpuPrevStats.Guest)},
			{"IRQ", fmt.Sprintf("%.2f%%", m.cpuStats.Irq), delta(m.cpuStats.Irq, m.cpuPrevStats.Irq)},
			{"SoftIRQ", fmt.Sprintf("%.2f%%", m.cpuStats.Softirq), delta(m.cpuStats.Softirq, m.cpuPrevStats.Softirq)},
			{"IOWait", fmt.Sprintf("%.2f%%", m.cpuStats.Iowait), delta(m.cpuStats.Iowait, m.cpuPrevStats.Iowait)},
		}

		m.cpuTable.SetRows(cpuRows)

		// Update RAM table information
		memRows := []table.Row{
			{"Total", fmt.Sprintf("%s", getByteMagnitude(m.memory.Total))},
			{"Used", fmt.Sprintf("%s", getByteMagnitude(m.memory.Used))},
			{"Available", fmt.Sprintf("%s", getByteMagnitude(m.memory.Available))},
			{"Free", fmt.Sprintf("%s", getByteMagnitude(m.memory.Free))},
			{"Buffers", fmt.Sprintf("%s", getByteMagnitude(m.memory.Buffers))},
			{"Cached", fmt.Sprintf("%s", getByteMagnitude(m.memory.Cached))},
		}

		m.memTable.SetRows(memRows)

		// Update Running Processes table information
		procRows := []table.Row{}
		for _, p := range processes {
			row := table.Row{
				fmt.Sprintf("%d", p.PID),
				p.Name,
				fmt.Sprint(p.Status),
				fmt.Sprint(p.Runtime),
				fmt.Sprintf("%s", getByteMagnitude(p.Memory)),
				fmt.Sprintf("%.2f%%", p.CPU),
			}
			procRows = append(procRows, row)
		}

		m.procTable.SetRows(procRows)

		// Update Disk table information
		diskRows := []table.Row{}
		for _, d := range m.disk {
			row := table.Row{
				fmt.Sprint(d.Partition.Mountpoint),
				d.Partition.Fstype,
				fmt.Sprintf("%s", getByteMagnitude(d.Total)),
				fmt.Sprintf("%s", getByteMagnitude(d.Used)),
				fmt.Sprintf("%s", getByteMagnitude(d.Free)),
			}
			diskRows = append(diskRows, row)
		}

		m.diskTable.SetRows(diskRows)

		return m, tea.Batch(cmd, tick())

	// Handle key pressing events
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Left):
			m.ActiveTab = max(m.ActiveTab-1, 0) // Decrement activeTab if possible
			return m, cmd
		case key.Matches(msg, m.keys.Right):
			m.ActiveTab = min(m.ActiveTab+1, len(m.tabs)-1) // Increment activeTab if possible
			return m, cmd
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll // Show full help message
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit // Quit program :(
		}

	}

	return m, nil
}

func (m model) View() string {

	if m.width == 0 || m.height == 0 {
		return "Loading...\nPress 'q' to quit."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'q' to quit.", m.err)
	}

	// Render category tabs
	// Assign style depending on active tab
	page := strings.Builder{}

	tabRow := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		if i == m.ActiveTab {
			tabRow[i] = activeTab.Render(t)
		} else {
			tabRow[i] = tab.Render(t)
		}
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tabRow...,
	)

	tabGap.MaxWidth(m.width)
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2))) // Gap line after tabs
	sep := tabGap.Render(strings.Repeat(" ", max(0, m.width)))                       // Bottom separator
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	page.WriteString(row + "\n\n")

	baseStyle.MaxWidth(m.width)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		page.String(),                         // Render category tabs
		m.renderTab(m.ActiveTab),              // Render active tab content
		fmt.Sprint(sep),                       // Render bottom separator
		baseStyle.Render(m.help.View(m.keys)), // Render help message with controls
	)

}
