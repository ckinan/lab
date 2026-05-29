import http from 'k6/http';
import { sleep } from 'k6';

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

export function steady() {
  http.post(TARGET, JSON.stringify({ delay_ms: 20 }), { headers });
}

export function latencySpike() {
  http.post(TARGET, JSON.stringify({ delay_ms: 800 }), { headers });
}

export function sawtoothLoad() {
  http.post(TARGET, JSON.stringify({ delay_ms: 50 }), { headers });
}

export function errorInjection() {
  http.post(TARGET, JSON.stringify({ fail: true }), { headers });
}

export function memoryHold() {
  http.post(TARGET, JSON.stringify({ mem_use_bytes: 10485760, mem_hold: true, delay_ms: 10 }), { headers });
}
