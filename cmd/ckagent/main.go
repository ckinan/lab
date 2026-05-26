package main

import (
"flag"
"log"
"net/http"
"time"

"github.com/VictoriaMetrics/metrics"
"github.com/ckinan/cktop/internal/adapters/proc"
)

func main() {
addr := flag.String("addr", ":9110", "HTTP listen address for /metrics")
flag.Parse()

memReader := proc.ProcMemoryReader{}
mem := &memCache{}
metrics.NewGauge(`ckagent_memory_total_bytes`, func() float64 { return float64(mem.get().Total) })
metrics.NewGauge(`ckagent_memory_free_bytes`, func() float64 { return float64(mem.get().Free) })
metrics.NewGauge(`ckagent_memory_available_bytes`, func() float64 { return float64(mem.get().Available) })
metrics.NewGauge(`ckagent_memory_used_bytes`, func() float64 { return float64(mem.get().Used) })
metrics.NewGauge(`ckagent_memory_buffers_bytes`, func() float64 { return float64(mem.get().Buffers) })
metrics.NewGauge(`ckagent_memory_cached_bytes`, func() float64 { return float64(mem.get().Cached) })
metrics.NewGauge(`ckagent_memory_shmem_bytes`, func() float64 { return float64(mem.get().Shmem) })
metrics.NewGauge(`ckagent_swap_total_bytes`, func() float64 { return float64(mem.get().SwapTotal) })
metrics.NewGauge(`ckagent_swap_free_bytes`, func() float64 { return float64(mem.get().SwapFree) })
metrics.NewGauge(`ckagent_swap_used_bytes`, func() float64 { return float64(mem.get().SwapUsed) })

cpu := &cpuCache{}
cpu.start(5*time.Second, proc.NewProcCPUReader())
metrics.NewGauge(`ckagent_cpu_usage_percent`, func() float64 { return cpu.get() })

cpuCores := newCPUCoresCache()
cpuCores.start(5*time.Second, proc.NewProcCPUCoresReader())

loadAvg := &loadAvgCache{}
metrics.NewGauge(`ckagent_load_avg_1m`, func() float64 { return loadAvg.get().Avg1m })
metrics.NewGauge(`ckagent_load_avg_5m`, func() float64 { return loadAvg.get().Avg5m })
metrics.NewGauge(`ckagent_load_avg_15m`, func() float64 { return loadAvg.get().Avg15m })
metrics.NewGauge(`ckagent_tasks_running`, func() float64 { return float64(loadAvg.get().Running) })
metrics.NewGauge(`ckagent_tasks_total`, func() float64 { return float64(loadAvg.get().Total) })

fileNR := &fileNRCache{}
metrics.NewGauge(`ckagent_fd_open`, func() float64 { return float64(fileNR.get().Open) })
metrics.NewGauge(`ckagent_fd_max`, func() float64 { return float64(fileNR.get().Max) })

sockStat := &sockStatCache{}
metrics.NewGauge(`ckagent_sockets_tcp_used`, func() float64 { return float64(sockStat.get().TCPUsed) })
metrics.NewGauge(`ckagent_sockets_tcp_orphan`, func() float64 { return float64(sockStat.get().TCPOrphan) })
metrics.NewGauge(`ckagent_sockets_tcp_timewait`, func() float64 { return float64(sockStat.get().TCPTimeWait) })
metrics.NewGauge(`ckagent_sockets_udp_used`, func() float64 { return float64(sockStat.get().UDPUsed) })
metrics.NewGauge(`ckagent_sockets_raw_used`, func() float64 { return float64(sockStat.get().RAWUsed) })

vmStat := &vmStatCache{}
metrics.NewGauge(`ckagent_vmstat_pgfault_total`, func() float64 { return float64(vmStat.get().PageFaults) })
metrics.NewGauge(`ckagent_vmstat_pgmajfault_total`, func() float64 { return float64(vmStat.get().MajorPageFaults) })
metrics.NewGauge(`ckagent_vmstat_pswpin_total`, func() float64 { return float64(vmStat.get().SwapIn) })
metrics.NewGauge(`ckagent_vmstat_pswpout_total`, func() float64 { return float64(vmStat.get().SwapOut) })
metrics.NewGauge(`ckagent_vmstat_pgpgin_total`, func() float64 { return float64(vmStat.get().PageIn) })
metrics.NewGauge(`ckagent_vmstat_pgpgout_total`, func() float64 { return float64(vmStat.get().PageOut) })

disk := newDiskCache()
netDev := newNetDevCache()
svc := newServiceCache()
svc.start(5*time.Second, proc.NewProcServiceStatsReader())

uptime := &uptimeCache{}
metrics.NewGauge(`ckagent_system_uptime_seconds`, func() float64 { return uptime.get() })

apt := &aptCache{}
metrics.NewGauge(`ckagent_apt_last_update_timestamp_seconds`, func() float64 { return float64(apt.get().LastUpdateUnix) })
metrics.NewGauge(`ckagent_apt_last_upgrade_timestamp_seconds`, func() float64 { return float64(apt.get().LastUpgradeUnix) })

psi := &psiCache{}
metrics.NewGauge(`ckagent_pressure_cpu_some_avg10`, func() float64 { return psi.get().CPU.Some.Avg10 })
metrics.NewGauge(`ckagent_pressure_cpu_some_avg60`, func() float64 { return psi.get().CPU.Some.Avg60 })
metrics.NewGauge(`ckagent_pressure_cpu_some_avg300`, func() float64 { return psi.get().CPU.Some.Avg300 })
metrics.NewGauge(`ckagent_pressure_memory_some_avg10`, func() float64 { return psi.get().Memory.Some.Avg10 })
metrics.NewGauge(`ckagent_pressure_memory_some_avg60`, func() float64 { return psi.get().Memory.Some.Avg60 })
metrics.NewGauge(`ckagent_pressure_memory_some_avg300`, func() float64 { return psi.get().Memory.Some.Avg300 })
metrics.NewGauge(`ckagent_pressure_memory_full_avg10`, func() float64 { return psi.get().Memory.Full.Avg10 })
metrics.NewGauge(`ckagent_pressure_memory_full_avg60`, func() float64 { return psi.get().Memory.Full.Avg60 })
metrics.NewGauge(`ckagent_pressure_memory_full_avg300`, func() float64 { return psi.get().Memory.Full.Avg300 })
metrics.NewGauge(`ckagent_pressure_io_some_avg10`, func() float64 { return psi.get().IO.Some.Avg10 })
metrics.NewGauge(`ckagent_pressure_io_some_avg60`, func() float64 { return psi.get().IO.Some.Avg60 })
metrics.NewGauge(`ckagent_pressure_io_some_avg300`, func() float64 { return psi.get().IO.Some.Avg300 })
metrics.NewGauge(`ckagent_pressure_io_full_avg10`, func() float64 { return psi.get().IO.Full.Avg10 })
metrics.NewGauge(`ckagent_pressure_io_full_avg60`, func() float64 { return psi.get().IO.Full.Avg60 })
metrics.NewGauge(`ckagent_pressure_io_full_avg300`, func() float64 { return psi.get().IO.Full.Avg300 })

cgroupPSI := newCgroupPSICache()

http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
mem.refresh(memReader)
loadAvg.refresh(proc.ProcLoadAvgReader{})
fileNR.refresh(proc.ProcFileNRReader{})
sockStat.refresh(proc.ProcSockStatReader{})
vmStat.refresh(proc.ProcVMStatReader{})
disk.refresh(proc.ProcDiskStatsReader{})
netDev.refresh(proc.ProcNetDevReader{})
uptime.refresh(proc.ProcUptimeReader{})
apt.refresh(proc.AptReader{})
psi.refresh(proc.ProcPSIReader{})
cgroupPSI.refresh(proc.CgroupPSIReader{})
metrics.WritePrometheus(w, false)
})

log.Printf("ckagent listening on %s", *addr)
log.Fatal(http.ListenAndServe(*addr, nil))
}
