package domain

type MemoryReader interface {
	ReadMemory() (Memory, error)
}

type ProcessReader interface {
	ReadProcesses() ([]Process, error)
}

type CPUReader interface {
	ReadCPU() (float64, error)
}

type CPUCoresReader interface {
	ReadCPUCores() (map[string]float64, error)
}

type LoadAvgReader interface {
	ReadLoadAvg() (LoadAvg, error)
}

type FileNRReader interface {
	ReadFileNR() (FileNR, error)
}

type SockStatReader interface {
	ReadSockStat() (SockStat, error)
}

type VMStatReader interface {
	ReadVMStat() (VMStat, error)
}

type DiskStatsReader interface {
	ReadDiskStats() ([]DiskStat, error)
}

type NetDevReader interface {
	ReadNetDev() ([]NetDev, error)
}

type ServiceStatsReader interface {
	ReadServiceStats() ([]ServiceStat, error)
}
