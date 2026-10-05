package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IncrViews counts one view of a product page. INCR is a single atomic command,
// so concurrent callers never lose an update, and no lock is needed.
func IncrViews(ctx context.Context, rdb redis.Cmdable, prefix string, id int64) (int64, error) {
	return rdb.Incr(ctx, fmt.Sprintf("%sviews:%d", prefix, id)).Result()
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
	n, err := fixedWindow.Run(ctx, rdb, []string{key}, window.Milliseconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("rate limit %s: %w", key, err)
	}
	return n <= limit, nil
}
