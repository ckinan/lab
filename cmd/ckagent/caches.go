package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/ckinan/lab/internal/adapters/proc"
	"github.com/ckinan/lab/internal/domain"
)

type memCache struct {
	mu  sync.RWMutex
	mem domain.Memory
}

func (c *memCache) refresh(r proc.ProcMemoryReader) {
	m, err := r.ReadMemory()
	if err != nil {
		log.Printf("read memory: %v", err)
		return
	}
	c.mu.Lock()
	c.mem = m
	c.mu.Unlock()
}

func (c *memCache) get() domain.Memory {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mem
}

type cpuCache struct {
	mu    sync.RWMutex
	value float64
}

func (c *cpuCache) start(interval time.Duration, r *proc.ProcCPUReader) {
	// Seed the baseline sample so the first tick produces a real delta.
	if _, err := r.ReadCPU(); err != nil {
		log.Printf("seed cpu baseline: %v", err)
	}
	go func() {
		t := time.NewTicker(interval)
		for range t.C {
			v, err := r.ReadCPU()
			if err != nil {
				log.Printf("read cpu: %v", err)
				continue
			}
			c.mu.Lock()
			c.value = v
			c.mu.Unlock()
		}
	}()
}

func (c *cpuCache) get() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

type cpuCoresCache struct {
	mu    sync.RWMutex
	cores map[string]float64
}

func newCPUCoresCache() *cpuCoresCache {
	return &cpuCoresCache{cores: make(map[string]float64)}
}

func (c *cpuCoresCache) start(interval time.Duration, r *proc.ProcCPUCoresReader) {
	if _, err := r.ReadCPUCores(); err != nil {
		log.Printf("seed cpu cores baseline: %v", err)
	}
	go func() {
		t := time.NewTicker(interval)
		for range t.C {
			vals, err := r.ReadCPUCores()
			if err != nil {
				log.Printf("read cpu cores: %v", err)
				continue
			}
			c.mu.Lock()
			c.cores = vals
			c.mu.Unlock()
			for core := range vals {
				coreName := core
				metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_cpu_core_usage_percent{core=%q}`, coreName), func() float64 {
					return c.getCore(coreName)
				})
			}
		}
	}()
}

func (c *cpuCoresCache) getCore(core string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cores[core]
}

type loadAvgCache struct {
	mu  sync.RWMutex
	val domain.LoadAvg
}

func (c *loadAvgCache) refresh(r proc.ProcLoadAvgReader) {
	v, err := r.ReadLoadAvg()
	if err != nil {
		log.Printf("read loadavg: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *loadAvgCache) get() domain.LoadAvg {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type fileNRCache struct {
	mu  sync.RWMutex
	val domain.FileNR
}

func (c *fileNRCache) refresh(r proc.ProcFileNRReader) {
	v, err := r.ReadFileNR()
	if err != nil {
		log.Printf("read file-nr: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *fileNRCache) get() domain.FileNR {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type sockStatCache struct {
	mu  sync.RWMutex
	val domain.SockStat
}

func (c *sockStatCache) refresh(r proc.ProcSockStatReader) {
	v, err := r.ReadSockStat()
	if err != nil {
		log.Printf("read sockstat: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *sockStatCache) get() domain.SockStat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type vmStatCache struct {
	mu  sync.RWMutex
	val domain.VMStat
}

func (c *vmStatCache) refresh(r proc.ProcVMStatReader) {
	v, err := r.ReadVMStat()
	if err != nil {
		log.Printf("read vmstat: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *vmStatCache) get() domain.VMStat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type diskCache struct {
	mu   sync.RWMutex
	devs map[string]domain.DiskStat
}

func newDiskCache() *diskCache {
	return &diskCache{devs: make(map[string]domain.DiskStat)}
}

func (c *diskCache) refresh(r proc.ProcDiskStatsReader) {
	stats, err := r.ReadDiskStats()
	if err != nil {
		log.Printf("read diskstats: %v", err)
		return
	}
	c.mu.Lock()
	for _, s := range stats {
		c.devs[s.Device] = s
	}
	c.mu.Unlock()
	for _, s := range stats {
		dev := s.Device
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_disk_reads_total{device=%q}`, dev), func() float64 {
			return float64(c.getDev(dev).ReadsTotal)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_disk_writes_total{device=%q}`, dev), func() float64 {
			return float64(c.getDev(dev).WritesTotal)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_disk_read_bytes_total{device=%q}`, dev), func() float64 {
			return float64(c.getDev(dev).ReadBytes)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_disk_write_bytes_total{device=%q}`, dev), func() float64 {
			return float64(c.getDev(dev).WriteBytes)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_disk_io_time_ms_total{device=%q}`, dev), func() float64 {
			return float64(c.getDev(dev).IOTimeMs)
		})
	}
}

func (c *diskCache) getDev(device string) domain.DiskStat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.devs[device]
}

type netDevCache struct {
	mu     sync.RWMutex
	ifaces map[string]domain.NetDev
}

func newNetDevCache() *netDevCache {
	return &netDevCache{ifaces: make(map[string]domain.NetDev)}
}

func (c *netDevCache) refresh(r proc.ProcNetDevReader) {
	devs, err := r.ReadNetDev()
	if err != nil {
		log.Printf("read netdev: %v", err)
		return
	}
	c.mu.Lock()
	for _, d := range devs {
		c.ifaces[d.Interface] = d
	}
	c.mu.Unlock()
	for _, d := range devs {
		iface := d.Interface
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_rx_bytes_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).RxBytes)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_tx_bytes_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).TxBytes)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_rx_packets_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).RxPackets)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_tx_packets_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).TxPackets)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_rx_errors_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).RxErrors)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_tx_errors_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).TxErrors)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_rx_drops_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).RxDrops)
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_net_tx_drops_total{interface=%q}`, iface), func() float64 {
			return float64(c.getIface(iface).TxDrops)
		})
	}
}

func (c *netDevCache) getIface(iface string) domain.NetDev {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ifaces[iface]
}

type psiCache struct {
	mu  sync.RWMutex
	val domain.SystemPSI
}

func (c *psiCache) refresh(r proc.ProcPSIReader) {
	v, err := r.ReadPSI()
	if err != nil {
		log.Printf("read psi: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *psiCache) get() domain.SystemPSI {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type cgroupPSICache struct {
	mu       sync.RWMutex
	services map[string]domain.ServicePSI
}

func newCgroupPSICache() *cgroupPSICache {
	return &cgroupPSICache{services: make(map[string]domain.ServicePSI)}
}

func (c *cgroupPSICache) refresh(r proc.CgroupPSIReader) {
	stats, err := r.ReadServicePSI()
	if err != nil {
		log.Printf("read cgroup psi: %v", err)
		return
	}
	c.mu.Lock()
	c.services = make(map[string]domain.ServicePSI, len(stats))
	for _, s := range stats {
		c.services[s.Unit] = s
	}
	c.mu.Unlock()
	for _, s := range stats {
		unit := s.Unit
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_pressure_cpu_some_avg10{unit=%q}`, unit), func() float64 {
			return c.getUnit(unit).CPU.Some.Avg10
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_pressure_memory_some_avg10{unit=%q}`, unit), func() float64 {
			return c.getUnit(unit).Memory.Some.Avg10
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_pressure_memory_full_avg10{unit=%q}`, unit), func() float64 {
			return c.getUnit(unit).Memory.Full.Avg10
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_pressure_io_some_avg10{unit=%q}`, unit), func() float64 {
			return c.getUnit(unit).IO.Some.Avg10
		})
		metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_pressure_io_full_avg10{unit=%q}`, unit), func() float64 {
			return c.getUnit(unit).IO.Full.Avg10
		})
	}
}

func (c *cgroupPSICache) getUnit(unit string) domain.ServicePSI {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.services[unit]
}

type uptimeCache struct {
	mu  sync.RWMutex
	val float64
}

func (c *uptimeCache) refresh(r proc.ProcUptimeReader) {
	v, err := r.ReadUptime()
	if err != nil {
		log.Printf("read uptime: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *uptimeCache) get() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type aptCache struct {
	mu  sync.RWMutex
	val domain.AptInfo
}

func (c *aptCache) refresh(r proc.AptReader) {
	v, err := r.ReadAptInfo()
	if err != nil {
		log.Printf("read apt info: %v", err)
		return
	}
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}

func (c *aptCache) get() domain.AptInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

type serviceCache struct {
	mu       sync.RWMutex
	services map[string]domain.ServiceStat
}

func newServiceCache() *serviceCache {
	return &serviceCache{services: make(map[string]domain.ServiceStat)}
}

func (c *serviceCache) start(interval time.Duration, r *proc.ProcServiceStatsReader) {
	if _, err := r.ReadServiceStats(); err != nil {
		log.Printf("seed service stats baseline: %v", err)
	}
	go func() {
		t := time.NewTicker(interval)
		for range t.C {
			stats, err := r.ReadServiceStats()
			if err != nil {
				log.Printf("read service stats: %v", err)
				continue
			}
			c.mu.Lock()
			c.services = make(map[string]domain.ServiceStat, len(stats))
			for _, s := range stats {
				c.services[s.Unit] = s
			}
			c.mu.Unlock()
			for _, s := range stats {
				unit := s.Unit
				metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_rss_bytes{unit=%q}`, unit), func() float64 {
					return float64(c.getUnit(unit).RSSBytes)
				})
				metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_cpu_percent{unit=%q}`, unit), func() float64 {
					return c.getUnit(unit).CPUPercent
				})
				metrics.GetOrCreateGauge(fmt.Sprintf(`ckagent_service_process_count{unit=%q}`, unit), func() float64 {
					return float64(c.getUnit(unit).ProcessCount)
				})
			}
		}
	}()
}

func (c *serviceCache) getUnit(unit string) domain.ServiceStat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.services[unit]
}
