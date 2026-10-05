// A load test for shorty (chapter 6), for k6 v2.
//
//   k6 run -e BASE_URL=http://127.0.0.1:8080 -e RATE=200 loadtest/shorty.js
//
// Open model: constant-arrival-rate starts RATE iterations per second whether or not the
// previous ones finished, the way independent users arrive. A slow server then shows up as
// higher latency and dropped iterations, not as a quietly lower request rate.
import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080';
const RATE = Number(__ENV.RATE || 200); // resolves per second in the measured phase
const LINKS = 100; // short codes created in setup and resolved during the test

export const options = {
  // The redirect is the answer we measure. Following it would load the target site and time
  // someone else's server.
  maxRedirects: 0,
  // Bodies are not needed for the checks below; discarding them saves memory in k6, not in shorty.
  discardResponseBodies: true,
  scenarios: {
    // Warm-up: same traffic at a quarter of the rate. Its samples carry phase:warmup, and no
    // threshold looks at them.
    warmup: {
      executor: 'constant-arrival-rate',
      exec: 'resolve',
      rate: Math.max(1, Math.floor(RATE / 4)),
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 10,
      maxVUs: 50,
      tags: { phase: 'warmup' },
    },
    resolve: {
      executor: 'constant-arrival-rate',
      exec: 'resolve',
      startTime: '10s',
      rate: RATE,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 20,
      maxVUs: 200,
      tags: { phase: 'measure' },
    },
    create: {
      executor: 'constant-arrival-rate',
      exec: 'create',
      startTime: '10s',
      rate: Math.max(1, Math.floor(RATE / 10)),
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 5,
      maxVUs: 50,
      tags: { phase: 'measure' },
    },
  },
  // Thresholds turn the run into a test: k6 exits non-zero when one fails.
  // These numbers are a starting point for a laptop; set yours from your own first run.
  thresholds: {
    'http_req_duration{phase:measure,name:resolve}': ['p(95)<25', 'p(99)<100'],
    'http_req_duration{phase:measure,name:create}': ['p(95)<50'],
    // Stop early if a quarter of the requests fail: there is nothing left to measure.
    'http_req_failed{phase:measure}': [
      { threshold: 'rate<0.01' },
      { threshold: 'rate<0.25', abortOnFail: true, delayAbortEval: '5s' },
    ],
    'checks{phase:measure}': ['rate>0.99'],
    // An iteration that could not start (no free VU) is load the server never saw.
    'dropped_iterations{scenario:resolve}': ['count==0'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
};

// setup runs once, before the scenarios; what it returns is passed to every iteration.
export function setup() {
  const codes = [];
  for (let i = 0; i < LINKS; i++) {
    const code = `k6-${Date.now()}-${i}`;
    const res = http.post(
      `${BASE_URL}/v1/links`,
      JSON.stringify({ url: `https://example.com/page/${i}`, code }),
      { headers: { 'Content-Type': 'application/json' }, tags: { name: 'setup' } },
    );
    if (res.status !== 201) {
      throw new Error(`setup: POST /v1/links returned ${res.status}; is shorty running at ${BASE_URL}?`);
    }
    codes.push(code);
  }
  return { codes };
}

export function resolve(data) {
  const code = data.codes[Math.floor(Math.random() * data.codes.length)];
  // The name tag groups 100 different URLs into one series; without it, every URL would be its
  // own time series in the results.
  const res = http.get(`${BASE_URL}/${code}`, { tags: { name: 'resolve' } });
  check(res, {
    'resolve: 302': (r) => r.status === 302,
    'resolve: has Location': (r) => (r.headers.Location || '').startsWith('https://example.com/'),
  });
}

export function create() {
  const res = http.post(
    `${BASE_URL}/v1/links`,
    JSON.stringify({ url: 'https://example.com/created-under-load' }),
    { headers: { 'Content-Type': 'application/json' }, tags: { name: 'create' } },
  );
  check(res, { 'create: 201': (r) => r.status === 201 });
}
