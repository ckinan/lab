import http from 'k6/http';

const TARGET = __ENV.CKAPI_URL || 'http://localhost:9120/work';

export const options = {
  scenarios: {
    steady: {
      executor: 'constant-arrival-rate',
      rate: 30,
      timeUnit: '1s',
      duration: '24h',
      preAllocatedVUs: 10,
      tags: { workload: 'steady' },
      exec: 'steady',
    },
    latency_spike: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      timeUnit: '1s',
      stages: [
        { duration: '4m30s', target: 20 },
        { duration: '30s',   target: 150 },
        { duration: '5m',    target: 20 },
        { duration: '14h44m', target: 20 },
      ],
      preAllocatedVUs: 120,
      maxVUs: 160,
      tags: { workload: 'latency-spike' },
      exec: 'latencySpike',
    },
    sawtooth_load: {
      executor: 'ramping-arrival-rate',
      startRate: 5,
      timeUnit: '1s',
      stages: [
        { duration: '1m',   target: 100 },
        { duration: '1m',   target: 5 },
        { duration: '1m',   target: 100 },
        { duration: '1m',   target: 5 },
        { duration: '1m',   target: 100 },
        { duration: '1m',   target: 5 },
        { duration: '23h54m', target: 5 },
      ],
      preAllocatedVUs: 30,
      tags: { workload: 'sawtooth-load' },
      exec: 'sawtoothLoad',
    },
    error_injection: {
      executor: 'constant-arrival-rate',
      rate: 10,
      timeUnit: '1s',
      duration: '24h',
      preAllocatedVUs: 5,
      tags: { workload: 'error-injection' },
      exec: 'errorInjection',
    },
    memory_hold: {
      executor: 'constant-arrival-rate',
      rate: 5,
      timeUnit: '1s',
      duration: '24h',
      preAllocatedVUs: 5,
      tags: { workload: 'memory-hold' },
      exec: 'memoryHold',
    },
  },
};

const headers = { 'Content-Type': 'application/json' };

function randInt(min, max) {
  return min + Math.floor(Math.random() * (max - min + 1));
}

// Steady baseline: light delay, occasional small CPU burn, no memory pressure.
export function steady() {
  const payload = { delay_ms: randInt(10, 60) };
  if (Math.random() < 0.3) payload.cpu_burn_ms = randInt(2, 10);
  http.post(TARGET, JSON.stringify(payload), { headers });
}

// Latency spike: high jitter delay, simulates slow downstream.
export function latencySpike() {
  http.post(TARGET, JSON.stringify({ delay_ms: randInt(500, 1200) }), { headers });
}

// Sawtooth load: oscillating RPS, variable delay and occasional memory allocations.
export function sawtoothLoad() {
  const payload = { delay_ms: randInt(5, 80) };
  if (Math.random() < 0.4) payload.mem_use_bytes = randInt(512 * 1024, 4 * 1024 * 1024);
  http.post(TARGET, JSON.stringify(payload), { headers });
}

// Error injection: mixed success/failure with varied CPU work on errors.
export function errorInjection() {
  const fail = Math.random() < 0.4;
  const payload = { delay_ms: randInt(5, 30) };
  if (fail) {
    payload.fail = true;
    payload.cpu_burn_ms = randInt(5, 20);
  }
  http.post(TARGET, JSON.stringify(payload), { headers });
}

// Memory hold: retain allocations across requests, varied size and hold duration.
export function memoryHold() {
  const memBytes = randInt(2 * 1024 * 1024, 20 * 1024 * 1024);
  http.post(TARGET, JSON.stringify({ mem_use_bytes: memBytes, mem_hold: true, delay_ms: randInt(5, 50) }), { headers });
}
