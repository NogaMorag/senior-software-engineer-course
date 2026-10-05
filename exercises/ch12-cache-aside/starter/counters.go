package catalog

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// IncrViews counts one view of a product page. INCR is a single atomic command,
// so concurrent callers never lose an update, and no lock is needed.
func IncrViews(ctx context.Context, rdb redis.Cmdable, prefix string, id int64) (int64, error) {
	// TODO(day 32): one INCR on prefix + "views:" + id.
	panic("TODO(day 32): IncrViews")
}

// fixedWindow increments a counter and starts its window on the first call. Run as
// one script, the two commands cannot be separated by a crash or by another client.
var fixedWindow = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return n
`)

// Allow is a fixed-window rate limiter, a preview of chapter 16: at most limit
// calls per window for key, counted in Redis so that every replica shares it.
func Allow(ctx context.Context, rdb redis.Scripter, key string, limit int64, window time.Duration) (bool, error) {
	// TODO(day 32): run fixedWindow with KEYS = [key] and ARGV = [window in milliseconds],
	// read the result as an int64 n, and allow when n <= limit.
	panic("TODO(day 32): Allow")
}
