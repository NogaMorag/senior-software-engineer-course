// Spec N1 (milestone M9): 99% of ingest requests return 2xx in under 200 ms, in the project's load
// profile, on your machine. For k6 v2; scripts/load.sh ingest starts the stack and passes the key.
//
//   k6 run -e BASE_URL=http://127.0.0.1:18080 -e API_KEY=hk_... -e RATE=100 loadtest/ingest.js
//
// The load profile: RATE events per second (default 100) from one tenant whose own limit is far above
// it, each with a new Idempotency-Key, for 60 s (DURATION) after a 10 s warm-up at a quarter of the rate.
// The tenant has an endpoint, so relay, dispatcher and workers run behind the API the whole time.
//
// Open model: constant-arrival-rate starts RATE iterations per second whether or not the previous ones
// finished. A slow API then shows up as latency and dropped iterations, not as a quietly lower rate.
import http from 'k6/http';
import { check } from 'k6';
import exec from 'k6/execution';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:18080';
const API_KEY = __ENV.API_KEY;
const RATE = Number(__ENV.RATE || 100);
const DURATION = __ENV.DURATION || '60s'; // the measured phase; scripts/chaos.sh makes it longer
const RUN = `${Date.now()}`; // makes every Idempotency-Key of this run new

export const options = {
  discardResponseBodies: true,
  scenarios: {
    // Same traffic at a quarter of the rate. Its samples carry phase:warmup and no threshold reads them.
    warmup: {
      executor: 'constant-arrival-rate',
      rate: Math.max(1, Math.floor(RATE / 4)),
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 10,
      maxVUs: 50,
      tags: { phase: 'warmup' },
    },
    ingest: {
      executor: 'constant-arrival-rate',
      startTime: '10s',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      // In-flight requests are rate times latency: 100 per second at 200 ms is 20 busy VUs. Allocate for
      // half a second, so a short stall is measured as latency instead of dropped while k6 adds VUs.
      preAllocatedVUs: Math.max(20, Math.ceil(RATE / 2)),
      maxVUs: Math.max(200, RATE * 2),
      tags: { phase: 'measure' },
    },
  },
  thresholds: {
    // N1 itself: the 2xx answers of the measured phase.
    'http_req_duration{phase:measure,expected_response:true}': ['p(99)<200'],
    'http_req_failed{phase:measure}': [
      { threshold: 'rate<0.01' },
      { threshold: 'rate<0.25', abortOnFail: true, delayAbortEval: '5s' },
    ],
    'checks{phase:measure}': ['rate>0.99'],
    // An iteration that could not start is load the API never saw: the run would flatter it.
    'dropped_iterations{scenario:ingest}': ['count==0'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
};

export function setup() {
  if (!API_KEY) {
    throw new Error('set API_KEY (scripts/load.sh ingest creates a tenant and passes it)');
  }
}

export default function () {
  const n = exec.scenario.iterationInTest;
  const res = http.post(
    `${BASE_URL}/v1/events`,
    JSON.stringify({ type: 'load.ingest', payload: { n } }),
    {
      headers: {
        Authorization: `Bearer ${API_KEY}`,
        'Content-Type': 'application/json',
        'Idempotency-Key': `k6-${RUN}-${exec.scenario.name}-${n}`,
      },
      tags: { name: 'ingest' },
    },
  );
  check(res, { 'ingest: 202': (r) => r.status === 202 });
}
