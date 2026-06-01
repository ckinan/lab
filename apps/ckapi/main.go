package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

var (
	memHeapAlloc  atomic.Uint64
	memHeapInuse  atomic.Uint64
	memHeapSys    atomic.Uint64
	memSys        atomic.Uint64
	memNumGC      atomic.Uint64
)

func startMemStats(ctx context.Context) {
	metrics.NewGauge(`go_memstats_heap_alloc_bytes`, func() float64 { return float64(memHeapAlloc.Load()) })
	metrics.NewGauge(`go_memstats_heap_inuse_bytes`, func() float64 { return float64(memHeapInuse.Load()) })
	metrics.NewGauge(`go_memstats_heap_sys_bytes`, func() float64 { return float64(memHeapSys.Load()) })
	metrics.NewGauge(`go_memstats_sys_bytes`, func() float64 { return float64(memSys.Load()) })
	metrics.NewGauge(`go_memstats_gc_runs_total`, func() float64 { return float64(memNumGC.Load()) })

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		var ms runtime.MemStats
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runtime.ReadMemStats(&ms)
				memHeapAlloc.Store(ms.HeapAlloc)
				memHeapInuse.Store(ms.HeapInuse)
				memHeapSys.Store(ms.HeapSys)
				memSys.Store(ms.Sys)
				memNumGC.Store(uint64(ms.NumGC))
			}
		}
	}()
}

func main() {
	addr := flag.String("addr", ":9120", "HTTP listen address")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	startMemStats(ctx)

	state := &State{}
	srv := New(state)

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	httpSrv := &http.Server{Addr: *addr, Handler: mux}

	go func() {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		if err := httpSrv.Shutdown(shutCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("ckapi listening on %s", *addr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
