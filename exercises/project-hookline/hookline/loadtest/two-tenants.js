// Spec N5 (milestone M9): a tenant over its rate limit gets 429 with Retry-After; the others are not
// affected. For k6 v2; scripts/load.sh two-tenants creates both tenants with known limits.
//
//   k6 run -e BASE_URL=http://127.0.0.1:18080 -e NOISY_KEY=hk_... -e QUIET_KEY=hk_... loadtest/two-tenants.js
//
// The noisy tenant sends NOISY_RATE per second (default 100) against a limit of a fifth of that, set by
// scripts/load.sh; the quiet one sends QUIET_RATE (default 20), well inside its own limit. Every sample is
// tagged with its tenant, so the thresholds below can judge each one separately.
import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:18080';
const KEYS = { noisy: __ENV.NOISY_KEY, quiet: __ENV.QUIET_KEY };
const NOISY_RATE = Number(__ENV.NOISY_RATE || 100);
const QUIET_RATE = Number(__ENV.QUIET_RATE || 20);
const RUN = `${Date.now()}`;

const limited = new Counter('ingest_429'); // 429 answers, tagged with the tenant

function tenant(name, rate) {
  return {
    executor: 'constant-arrival-rate',
    exec: name,
    rate,
    timeUnit: '1s',
    duration: '40s',
    preAllocatedVUs: Math.max(10, Math.ceil(rate / 2)), // half a second of in-flight requests, as in ingest.js
    maxVUs: Math.max(100, rate * 2),
    tags: { tenant: name },
  };
}

export const options = {
  discardResponseBodies: true,
  scenarios: { noisy: tenant('noisy', NOISY_RATE), quiet: tenant('quiet', QUIET_RATE) },
  thresholds: {
    // The noisy tenant is limited, and every limited answer says when to come back.
    'ingest_429{tenant:noisy}': ['count>0'],
    'checks{tenant:noisy}': ['rate==1'],
    // The quiet tenant is not affected: never limited, every event accepted, and as fast as N1 asks.
    'ingest_429{tenant:quiet}': ['count==0'],
    'checks{tenant:quiet}': ['rate==1'],
    'http_req_duration{tenant:quiet}': ['p(99)<200'],
    'dropped_iterations': ['count==0'],
  },
  summaryTrendStats: ['avg', 'med', 'p(95)', 'p(99)', 'max'],
};

export function setup() {
  if (!KEYS.noisy || !KEYS.quiet) {
    throw new Error('set NOISY_KEY and QUIET_KEY (scripts/load.sh two-tenants creates both tenants)');
  }
}

function send(name) {
  const n = exec.scenario.iterationInTest;
  const res = http.post(
    `${BASE_URL}/v1/events`,
    JSON.stringify({ type: 'load.two_tenants', payload: { n } }),
    {
      headers: {
        Authorization: `Bearer ${KEYS[name]}`,
        'Content-Type': 'application/json',
        'Idempotency-Key': `k6-${RUN}-${name}-${n}`,
      },
      tags: { name: 'ingest' },
      // A 429 is the answer this test expects from the noisy tenant, not a failed request.
      responseCallback: http.expectedStatuses(202, 429),
    },
  );
  if (res.status === 429) {
    limited.add(1);
  }
  return res;
}

export function noisy() {
  check(send('noisy'), {
    'noisy: 202 or 429': (r) => r.status === 202 || r.status === 429,
    'noisy: a 429 has Retry-After in whole seconds': (r) =>
      r.status !== 429 || /^[1-9][0-9]*$/.test(r.headers['Retry-After'] || ''),
  });
}

export function quiet() {
  check(send('quiet'), { 'quiet: 202': (r) => r.status === 202 });
}
