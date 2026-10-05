-- Token bucket, one Redis hash per key, run atomically by Redis.
-- KEYS[1]  the bucket's hash (fields: tokens, ts)
-- ARGV[1]  rate in tokens per second
-- ARGV[2]  burst (bucket size)
-- ARGV[3]  tokens requested
-- Returns  {allowed (1 or 0), whole tokens remaining, retry-after in milliseconds}
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local n = tonumber(ARGV[3])

-- TODO(day 48), in order:
--  1. The clock: redis.call('TIME') returns {seconds, microseconds}; turn it into
--     milliseconds (Lua 5.1 prints 14 significant digits, so not microseconds).
--  2. HMGET the fields tokens and ts; a missing bucket is full (tokens = burst, ts = now).
--  3. Refill: tokens + (now - ts) * rate / 1000, never above burst, never for a negative gap.
--  4. Enough tokens: take n, allowed = 1. Otherwise retry_ms = ceil((n - tokens) * 1000 / rate).
--  5. HSET tokens and ts (as strings, with tostring), then PEXPIRE the key for
--     ceil(burst * 1000 / rate) + 1000 ms: an idle bucket is full again by then.
--  6. return {allowed, math.floor(tokens), retry_ms}
return redis.error_reply('TODO(day 48): tokenbucket.lua')
