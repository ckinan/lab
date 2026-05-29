package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/ckinan/lab/ckapi/internal/domain"
)

type Server struct {
	active atomic.Int64
	state  *domain.State
}

func New(state *domain.State) *Server {
	s := &Server{state: state}
	metrics.NewGauge(`ckapi_active_requests`, func() float64 {
		return float64(s.active.Load())
	})
	return s
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /work", s.work)
	mux.HandleFunc("POST /control", s.setControl)
	mux.HandleFunc("GET /control", s.getControl)
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {})
}

func (s *Server) work(w http.ResponseWriter, r *http.Request) {
	var req domain.BehaviorReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	req = merge(s.state.Defaults(), req)

	s.active.Add(1)
	defer s.active.Add(-1)

	start := time.Now()
	code := execute(r.Context(), req, s.state)
	dur := time.Since(start)

	label := fmt.Sprintf(`method="POST",path="/work",status_code=%q`, strconv.Itoa(code))
	metrics.GetOrCreateCounter(`ckapi_requests_total{` + label + `}`).Inc()
	metrics.GetOrCreateHistogram(`ckapi_request_duration_seconds{method="POST",path="/work"}`).Update(dur.Seconds())

	w.WriteHeader(code)
}

func (s *Server) setControl(w http.ResponseWriter, r *http.Request) {
	var req domain.BehaviorReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.state.SetDefaults(req)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getControl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.state.Defaults())
}

func merge(defaults, req domain.BehaviorReq) domain.BehaviorReq {
	if req.DelayMs == 0 {
		req.DelayMs = defaults.DelayMs
	}
	if req.CPUBurnMs == 0 {
		req.CPUBurnMs = defaults.CPUBurnMs
	}
	if req.MemBytes == 0 {
		req.MemBytes = defaults.MemBytes
	}
	if req.MemHold == nil {
		req.MemHold = defaults.MemHold
	}
	if req.Fail == nil {
		req.Fail = defaults.Fail
	}
	if req.StatusCode == 0 {
		req.StatusCode = defaults.StatusCode
	}
	return req
}

func execute(ctx context.Context, req domain.BehaviorReq, state *domain.State) int {
	if req.MemBytes > 0 && req.MemHold != nil && *req.MemHold {
		release := state.Allocate(req.MemBytes)
		defer release()
	}

	if req.DelayMs > 0 {
		select {
		case <-time.After(time.Duration(req.DelayMs) * time.Millisecond):
		case <-ctx.Done():
			return http.StatusGatewayTimeout
		}
	}

	if req.CPUBurnMs > 0 {
		deadline := time.Now().Add(time.Duration(req.CPUBurnMs) * time.Millisecond)
		for time.Now().Before(deadline) {
		}
		metrics.GetOrCreateCounter(`ckapi_cpu_burn_milliseconds_total`).Add(req.CPUBurnMs)
	}

	if req.MemBytes > 0 && (req.MemHold == nil || !*req.MemHold) {
		buf := make([]byte, req.MemBytes)
		domain.TouchPages(buf)
		runtime.KeepAlive(buf)
	}

	if req.Fail != nil && *req.Fail {
		if req.StatusCode > 0 {
			return req.StatusCode
		}
		return http.StatusInternalServerError
	}
	if req.StatusCode > 0 {
		return req.StatusCode
	}
	return http.StatusOK
}
