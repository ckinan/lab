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
