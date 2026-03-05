// k6 load test for REST API endpoints
// Usage: k6 run tests/load/k6-api.js
//
// Prerequisites: Server running at BASE_URL with JWT_SECRET configured
// Set AUTH_TOKEN env var or the test will attempt login

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const apiDuration = new Trend('api_duration');
const requestCount = new Counter('total_requests');

export const options = {
  stages: [
    { duration: '15s', target: 50 },    // Warm up
    { duration: '30s', target: 200 },   // Ramp to 200 concurrent
    { duration: '1m', target: 500 },    // Ramp to 500
    { duration: '2m', target: 1000 },   // Peak at 1000
    { duration: '30s', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<200'],   // p99 <200ms
    errors: ['rate<0.01'],              // Error rate <1%
    'http_req_duration{endpoint:health}': ['p(95)<50'],  // Health check <50ms
    'http_req_duration{endpoint:models}': ['p(95)<100'], // Model list <100ms
    'http_req_duration{endpoint:devices}': ['p(95)<100'], // Device list <100ms
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Get or create auth token
let authToken = __ENV.AUTH_TOKEN || '';

export function setup() {
  if (authToken) {
    return { token: authToken };
  }

  // Try to login
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: 'admin@fleetml.io',
    password: 'admin123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (loginRes.status === 200) {
    const body = JSON.parse(loginRes.body);
    return { token: body.access_token || body.token || '' };
  }

  // If login fails, try register
  const regRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    email: 'admin@fleetml.io',
    password: 'admin123',
    role: 'admin',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (regRes.status === 200 || regRes.status === 201) {
    const body = JSON.parse(regRes.body);
    return { token: body.access_token || body.token || '' };
  }

  return { token: '' };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
  };

  if (data.token) {
    headers['Authorization'] = `Bearer ${data.token}`;
  }

  const params = { headers };

  // Distribute load across endpoints by VU iteration
  const endpoint = __ITER % 5;

  switch (endpoint) {
    case 0:
      group('health', function () {
        const res = http.get(`${BASE_URL}/api/v1/health`, {
          tags: { endpoint: 'health' },
        });
        apiDuration.add(res.timings.duration);
        requestCount.add(1);
        const ok = check(res, {
          'health 200': (r) => r.status === 200,
          'health <50ms': (r) => r.timings.duration < 50,
        });
        errorRate.add(!ok);
      });
      break;

    case 1:
      group('list_models', function () {
        const res = http.get(`${BASE_URL}/api/v1/models`, {
          ...params,
          tags: { endpoint: 'models' },
        });
        apiDuration.add(res.timings.duration);
        requestCount.add(1);
        const ok = check(res, {
          'models 200 or 401': (r) => r.status === 200 || r.status === 401,
          'models <100ms': (r) => r.timings.duration < 100,
        });
        errorRate.add(!ok);
      });
      break;

    case 2:
      group('list_devices', function () {
        const res = http.get(`${BASE_URL}/api/v1/devices`, {
          ...params,
          tags: { endpoint: 'devices' },
        });
        apiDuration.add(res.timings.duration);
        requestCount.add(1);
        const ok = check(res, {
          'devices 200 or 401': (r) => r.status === 200 || r.status === 401,
          'devices <100ms': (r) => r.timings.duration < 100,
        });
        errorRate.add(!ok);
      });
      break;

    case 3:
      group('list_deployments', function () {
        const res = http.get(`${BASE_URL}/api/v1/deployments`, {
          ...params,
          tags: { endpoint: 'deployments' },
        });
        apiDuration.add(res.timings.duration);
        requestCount.add(1);
        const ok = check(res, {
          'deployments 200 or 401': (r) => r.status === 200 || r.status === 401,
          'deployments <200ms': (r) => r.timings.duration < 200,
        });
        errorRate.add(!ok);
      });
      break;

    case 4:
      group('heartbeat', function () {
        const payload = JSON.stringify({
          device_id: `load-device-${__VU}`,
          status: 'healthy',
          system: {
            cpu_percent: Math.random() * 80 + 10,
            gpu_percent: Math.random() * 60,
            ram_used_mb: Math.floor(Math.random() * 4096),
            disk_percent: Math.random() * 70 + 10,
            temperature_c: Math.random() * 30 + 30,
          },
        });

        const res = http.post(`${BASE_URL}/api/v1/heartbeat`, payload, {
          headers: { 'Content-Type': 'application/json' },
          tags: { endpoint: 'heartbeat' },
        });
        apiDuration.add(res.timings.duration);
        requestCount.add(1);
        const ok = check(res, {
          'heartbeat 200': (r) => r.status === 200,
          'heartbeat <200ms': (r) => r.timings.duration < 200,
        });
        errorRate.add(!ok);
      });
      break;
  }

  sleep(0.5 + Math.random() * 0.5); // 500-1000ms between requests
}

export function handleSummary(data) {
  return {
    'tests/load/api-results.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: '  ', enableColors: true }),
  };
}

function textSummary(data) {
  const metrics = data.metrics;
  let summary = '\n=== FleetML API Load Test Results ===\n\n';

  if (metrics.http_req_duration) {
    const dur = metrics.http_req_duration.values;
    summary += `Request Duration:\n`;
    summary += `  p50: ${dur['p(50)']?.toFixed(1) || 'N/A'}ms\n`;
    summary += `  p95: ${dur['p(95)']?.toFixed(1) || 'N/A'}ms\n`;
    summary += `  p99: ${dur['p(99)']?.toFixed(1) || 'N/A'}ms\n`;
  }

  if (metrics.errors) {
    summary += `Error Rate: ${(metrics.errors.values.rate * 100).toFixed(2)}%\n`;
  }

  if (metrics.total_requests) {
    summary += `Total Requests: ${metrics.total_requests.values.count}\n`;
  }

  return summary;
}
