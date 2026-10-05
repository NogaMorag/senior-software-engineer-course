// Spec N6 (milestone M10): above capacity, ingest sheds load with 503 instead of queueing without bound.
// For k6 v2; scripts/load.sh overload starts the API with a small HOOKLINE_MAX_IN_FLIGHT (MAX_IN_FLIGHT,
// default 4) and a tenant whose own rate limit is far above RATE, so the shedder, not the limiter,
// answers.
//
//   k6 run -e BASE_URL=http://127.0.0.1:18080 -e API_KEY=hk_... -e RATE=800 loadtest/overload.js
//
// What counts as passing: shedding really happened (some 503s), every answer was 202 or a 503 with
// Retry-After, the answers stayed fast (a queue that grows shows up as latency that grows), and the
// health route kept answering 200 throughout, because probes are never shed.
import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:18080';
const API_KEY = __ENV.API_KEY;
const RATE = Number(__ENV.RATE || 800);
const RUN = `${Date.now()}`;

const shed = new Counter('ingest_503');

export const options = {
  discardResponseBodies: true,
  scenarios: {
    flood: {
      executor: 'constant-arrival-rate',
      exec: 'ingest',
      rate: RATE,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: Math.max(50, Math.ceil(RATE / 2)),
      maxVUs: RATE * 2,
      tags: { name: 'ingest' },
    },
    probe: {
      executor: 'constant-arrival-rate',
      exec: 'health',
      rate: 5,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 2,
      maxVUs: 10,
      tags: { name: 'health' },
    },
  },
  thresholds: {
    ingest_503: ['count>0'],
    'checks{name:ingest}': ['rate==1'],
    'http_req_duration{name:ingest}': ['p(99)<1000'],
    'checks{name:health}': ['rate==1'],
  },
  summaryTrendStats: ['avg', 'med', 'p(95)', 'p(99)', 'max'],
};

export function setup() {
  if (!API_KEY) {
    throw new Error('set API_KEY (scripts/load.sh overload creates a tenant and passes it)');
  }
}

export function ingest() {
  const n = exec.scenario.iterationInTest;
  const res = http.post(
    `${BASE_URL}/v1/events`,
    JSON.stringify({ type: 'load.overload', payload: { n } }),
    {
      headers: {
        Authorization: `Bearer ${API_KEY}`,
        'Content-Type': 'application/json',
        'Idempotency-Key': `k6-${RUN}-${n}`,
      },
      // A 503 is the answer this test is looking for, not a failed request.
      responseCallback: http.expectedStatuses(202, 503),
    },
  );
  if (res.status === 503) {
    shed.add(1);
  }
  check(res, {
    'ingest: 202 or 503': (r) => r.status === 202 || r.status === 503,
    'ingest: a 503 has Retry-After': (r) => r.status !== 503 || /^[1-9][0-9]*$/.test(r.headers['Retry-After'] || ''),
  });
}

export function health() {
  check(http.get(`${BASE_URL}/healthz`), { 'health: 200': (r) => r.status === 200 });
}
