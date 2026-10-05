-- Token bucket, one Redis hash per key, run atomically by Redis.
-- KEYS[1]  the bucket's hash (fields: tokens, ts)
-- ARGV[1]  rate in tokens per second
-- ARGV[2]  burst (bucket size)
-- ARGV[3]  tokens requested
-- Returns  {allowed (1 or 0), whole tokens remaining, retry-after in milliseconds}
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local n = tonumber(ARGV[3])

-- One clock for every instance: Redis' own, in milliseconds. (Lua 5.1 prints numbers
-- with 14 significant digits, so microseconds would not survive the round trip.)
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)

local state = redis.call('HMGET', KEYS[1], 'tokens', 'ts')
local tokens = tonumber(state[1]) or burst
local ts = tonumber(state[2]) or now

-- Refill for the time since the last request, never above burst.
tokens = math.min(burst, tokens + math.max(0, now - ts) * rate / 1000)

local allowed = 0
local retry_ms = 0
if tokens >= n then
  tokens = tokens - n
  allowed = 1
else
  retry_ms = math.ceil((n - tokens) * 1000 / rate)
end

redis.call('HSET', KEYS[1], 'tokens', tostring(tokens), 'ts', tostring(now))
-- An idle bucket is full again after burst/rate seconds; after that the key is useless.
redis.call('PEXPIRE', KEYS[1], math.ceil(burst * 1000 / rate) + 1000)
return {allowed, math.floor(tokens), retry_ms}
