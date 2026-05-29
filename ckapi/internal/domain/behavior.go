package domain

import "sync"

type BehaviorReq struct {
	DelayMs    int    `json:"delay_ms"`
	CPUBurnMs  int    `json:"cpu_burn_ms"`
	MemBytes   int64  `json:"mem_use_bytes"`
	MemHold    *bool  `json:"mem_hold,omitempty"`
	Fail       *bool  `json:"fail,omitempty"`
	StatusCode int    `json:"status_code"`
}

type State struct {
	mu          sync.Mutex
	defaults    BehaviorReq
	nextID      int
	allocations map[int][]byte
}

func (s *State) SetDefaults(r BehaviorReq) {
	s.mu.Lock()
	s.defaults = r
	s.mu.Unlock()
}

func (s *State) Defaults() BehaviorReq {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.defaults
}

// Allocate adds n bytes to the held pool and returns a release function.
func (s *State) Allocate(n int64) func() {
	buf := make([]byte, n)
	TouchPages(buf)

	s.mu.Lock()
	if s.allocations == nil {
		s.allocations = make(map[int][]byte)
	}
	id := s.nextID
	s.nextID++
	s.allocations[id] = buf
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		delete(s.allocations, id)
		s.mu.Unlock()
	}
}

func TouchPages(b []byte) {
	for i := 0; i < len(b); i += 4096 {
		b[i] = 1
	}
}
