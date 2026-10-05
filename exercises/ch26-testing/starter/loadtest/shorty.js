// A load test for shorty (chapter 6), for k6 v2.
//
//   k6 run -e BASE_URL=http://127.0.0.1:8080 -e RATE=200 loadtest/shorty.js
//
// This skeleton runs, but it is a closed-model smoke test without thresholds: it measures
// whatever rate shorty happens to sustain and never fails. Day 81, steps 4 and 5 turn it into
// a load test.
import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080';

export const options = {
  // TODO: maxRedirects: 0, so a 302 is the answer instead of a fetch of the target site.
  // TODO: scenarios: warmup (rate RATE/4, 10s, tag phase:warmup), resolve (RATE, 30s, from
  //       startTime 10s, tag phase:measure) and create (RATE/10), all constant-arrival-rate.
  // TODO: thresholds on http_req_duration{phase:measure,name:resolve} (p95, p99),
  //       http_req_failed{phase:measure}, checks{phase:measure}, dropped_iterations.
  vus: 5,
  duration: '10s',
};

// setup runs once; what it returns is passed to every iteration.
export function setup() {
  const code = `k6-${Date.now()}`;
  const res = http.post(
    `${BASE_URL}/v1/links`,
    JSON.stringify({ url: 'https://example.com/page/0', code }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (res.status !== 201) {
    throw new Error(`setup: POST /v1/links returned ${res.status}; is shorty running at ${BASE_URL}?`);
  }
  return { codes: [code] };
}

export default function (data) {
  // TODO: tag the request with { name: 'resolve' } and pick a random code from data.codes.
  const res = http.get(`${BASE_URL}/${data.codes[0]}`);
  check(res, { 'status is 2xx or 3xx': (r) => r.status < 400 });
}
