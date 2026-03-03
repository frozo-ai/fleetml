// k6 load test for heartbeat endpoint
// Usage: k6 run tests/load/k6-heartbeat.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const heartbeatDuration = new Trend('heartbeat_duration');

export const options = {
  stages: [
    { duration: '30s', target: 100 },   // Ramp up to 100 concurrent
    { duration: '1m', target: 500 },    // Ramp to 500
    { duration: '2m', target: 1000 },   // Peak at 1000
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],  // 95th percentile <200ms
    errors: ['rate<0.01'],              // Error rate <1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const deviceId = `load-test-device-${__VU}-${__ITER}`;

  // Simulate heartbeat via REST fallback
  const payload = JSON.stringify({
    device_id: deviceId,
    status: 'healthy',
    system: {
      cpu_percent: Math.random() * 80 + 10,
      gpu_percent: Math.random() * 60,
      ram_used_mb: Math.floor(Math.random() * 4096),
      disk_percent: Math.random() * 70 + 10,
      temperature_c: Math.random() * 30 + 30,
      uptime_hours: Math.random() * 720,
    },
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const start = new Date();
  const res = http.post(`${BASE_URL}/api/v1/heartbeat`, payload, params);
  const duration = new Date() - start;

  heartbeatDuration.add(duration);

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response time <200ms': (r) => r.timings.duration < 200,
  });

  errorRate.add(!success);

  // Simulate heartbeat interval
  sleep(1);
}

export function handleSummary(data) {
  return {
    'tests/load/heartbeat-results.json': JSON.stringify(data, null, 2),
  };
}
