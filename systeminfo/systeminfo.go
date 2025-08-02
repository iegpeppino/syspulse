package systeminfo

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// Get general CPU Info
// Will be used in future implementations
func GetCPUinfo() (cpu.InfoStat, error) {

	cpuInfo, err := cpu.Info()
	if err != nil {
		return cpu.InfoStat{}, fmt.Errorf("unable to get CPU info: %w", err)
	}

	return cpuInfo[0], nil
}

// Returns percentual value of total CPU usage
func GetCPUPercent() (float64, error) {
	cpuPercentage, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return 0.0, err
	}

	return cpuPercentage[0], nil
}

// Return Cpu times spent on different processes
func GetCPUTimes() (*[]cpu.TimesStat, error) {

	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return &[]cpu.TimesStat{}, err
	}

	currTimes := cpuTimes[0]

	// Calculate combined CPU times
	totalTime := currTimes.Guest + currTimes.Idle + currTimes.Iowait + currTimes.Irq +
		currTimes.Nice + currTimes.Softirq + currTimes.Steal + currTimes.System + currTimes.User

	// Convert times to percentual values using totalTimes
	currTimes.Guest = (currTimes.Guest / totalTime) * 100     // Guest operating systems
	currTimes.Idle = (currTimes.Idle / totalTime) * 100       // Not actively executing tasks
	currTimes.Iowait = (currTimes.Iowait / totalTime) * 100   // I/O ops
	currTimes.Irq = (currTimes.Irq / totalTime) * 100         // Interrupt service
	currTimes.Nice = (currTimes.Nice / totalTime) * 100       // Low schedule priority processes
	currTimes.Softirq = (currTimes.Softirq / totalTime) * 100 // Software Interrupt Service
	currTimes.Steal = (currTimes.Steal / totalTime) * 100     // Cpu stolen by virtual environments
	currTimes.System = (currTimes.System / totalTime) * 100   // Executing OS's (kernel's) code
	currTimes.User = (currTimes.User / totalTime) * 100       // Executing programs' instructions within user space

	cpuTimes[0] = currTimes
	return &cpuTimes, nil

}

// Returns Memory usage statistics
func GetMEMLoad() (*mem.VirtualMemoryStat, error) {

	v, err := mem.VirtualMemory()
	if err != nil {
		return &mem.VirtualMemoryStat{}, err
	}

	return &mem.VirtualMemoryStat{
		Total:       v.Total,       // Total RAM amount on this system
		Available:   v.Available,   // RAM available to allocate
		Used:        v.Used,        // RAM used by programs
		UsedPercent: v.UsedPercent, // RAM used by programs in percentual value
		Free:        v.Free,        // Kernel's notion of free memory
		Buffers:     v.Buffers,
		Cached:      v.Cached,
	}, nil

}

// Disk partition stats struct
type DiskInfo struct {
	Partition disk.PartitionStat
	Fstype    string
	Total     uint64
	Free      uint64
	Used      uint64
}

func GetDISKUse() ([]DiskInfo, error) {

	partitions, err := disk.Partitions(true)
	var disks []DiskInfo
	if err != nil {
		return disks, fmt.Errorf("unable to get disk info: %w", err)
	}

	for _, p := range partitions {
		diskInfo := DiskInfo{}
		diskInfo.Partition = p
		// Use mountpoint to get disks
		// Virtual memory returns filepaths
		usageStat, err := disk.Usage(p.Mountpoint)
		if err != nil {
			return disks, fmt.Errorf("unable to get disk stats: %w", err)
		}
		diskInfo.Free = usageStat.Free
		diskInfo.Fstype = usageStat.Fstype
		diskInfo.Total = usageStat.Total
		diskInfo.Used = usageStat.Used

		disks = append(disks, diskInfo)
	}

	if len(disks) == 0 {
		return disks, errors.New("disks couldn't be found")
	}

	// Sort Disk by total capacity
	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Total > disks[j].Total
	})

	return disks, nil
}

// Running process stats struct
type ProcessInfo struct {
	PID     int32
	Name    string
	CPU     float64
	Memory  uint64
	Runtime string
	Status  []string
}

func GetProcessInfo(n int) ([]ProcessInfo, error) {

	processes, err := process.Processes()
	if err != nil {
		return []ProcessInfo{}, err
	}

	var processesInfo []ProcessInfo
	var procErr error

	for _, p := range processes {
		proc := ProcessInfo{}
		proc.PID = p.Pid

		proc.Name, err = p.Name()
		if err != nil {
			procErr = errors.Join(procErr, err)
			proc.Name = "N/A"
		}

		proc.Status, err = p.Status()
		if err != nil {
			procErr = errors.Join(procErr, err)
			proc.Status = []string{"Unknown"}
		}

		started, err := p.CreateTime()
		if err != nil {
			procErr = errors.Join(procErr, err)
			proc.Runtime = "N/A"
		}

		// Divide by 1000 since CreateTime() returns uint time in milliseconds
		runtime := time.Since(time.Unix(started/1000, 0)).Truncate(time.Second)
		proc.Runtime = runtime.String()

		cpuInfo, err := p.CPUPercent()
		if err != nil {
			procErr = errors.Join(procErr, err)
			proc.CPU = 0.0
		}

		proc.CPU = cpuInfo

		memoryInfo, err := p.MemoryInfo()
		// If for loop is not broken after a memoryInfo error
		// a runtime error occurs
		if err != nil {
			procErr = errors.Join(procErr, err)
			processesInfo = append(processesInfo, ProcessInfo{
				PID:     proc.PID,
				Name:    proc.Name,
				Status:  proc.Status,
				Runtime: proc.Runtime,
				Memory:  0.0,
				CPU:     0.0,
			})
			continue
		}

		proc.Memory = memoryInfo.RSS

		processesInfo = append(processesInfo, proc)
	}

	// Getting only "n" number of processes
	if len(processesInfo) > n {
		processesInfo = processesInfo[:n]
	}

	// Sorting processes by CPU usage
	sort.Slice(processesInfo, func(i, j int) bool {
		return processesInfo[i].CPU > processesInfo[j].CPU
	})

	return processesInfo, procErr
}
