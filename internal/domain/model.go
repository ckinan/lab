package domain

type Memory struct {
	Total     int64 `json:"total"`
	Free      int64 `json:"free"`
	Available int64 `json:"available"`
	Used      int64 `json:"used"`
	Buffers   int64 `json:"buffers"`
	Cached    int64 `json:"cached"`
	Shmem     int64 `json:"shmem"`
	SwapTotal int64 `json:"swap_total"`
	SwapFree  int64 `json:"swap_free"`
	SwapUsed  int64 `json:"swap_used"`
}

type Process struct {
	Pid       int
	Ppid      int
	Rss       int
	CPU       float64
	Cmdline   string
	Username  string
	IsKthread bool
}

type Snapshot struct {
	CPU       float64
	Memory    Memory
	Processes []Process
}

type LoadAvg struct {
	Avg1m   float64
	Avg5m   float64
	Avg15m  float64
	Running int
	Total   int
}

type FileNR struct {
	Open int64
	Max  int64
}

type SockStat struct {
	TCPUsed     int
	TCPOrphan   int
	TCPTimeWait int
	UDPUsed     int
	RAWUsed     int
}

type ServiceStat struct {
	Unit         string
	RSSBytes     int64
	CPUPercent   float64
	ProcessCount int
}

type VMStat struct {
	PageFaults      uint64
	MajorPageFaults uint64
	SwapIn          uint64
	SwapOut         uint64
	PageIn          uint64
	PageOut         uint64
}

type DiskStat struct {
	Device      string
	ReadsTotal  uint64
	WritesTotal uint64
	ReadBytes   uint64
	WriteBytes  uint64
	IOTimeMs    uint64
}

type NetDev struct {
	Interface string
	RxBytes   uint64
	RxPackets uint64
	RxErrors  uint64
	RxDrops   uint64
	TxBytes   uint64
	TxPackets uint64
	TxErrors  uint64
	TxDrops   uint64
}
