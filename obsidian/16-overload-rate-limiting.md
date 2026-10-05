---
direction: rtl
tags: [senior-course, book]
chapter: 16
---
# עומס: הגבלת קצב, לחץ חוזר והשלת עומס

- **תור תמיד קיים:** goroutine שמחכה, חיבור שמחכה; השאלה היא אם הוא חסום.
- **חוק ליטל:** בפנים שווה קצב כפול זמן; מגבילים את מה שבפנים כדי לחסום את הזמן.
- **רוויה:** זמן התגובה הוא זמן השירות חלקי אחד פחות הניצולת; ב־90 אחוז פי עשרה.
- **דלי אסימונים:** קצב ופרץ; מילוי עצל; זמן חזרה מעוגל למעלה; שעון מוזרק.
- **מבוזר:** סקריפט אטומי, שעון של Redis, מפתחות רק בארגומנטים, תפוגה לכל דלי.
- **מטמון סקריפטים:** נדיף; תשובת NOSCRIPT, והספרייה שולחת את הסקריפט המלא.
- **קודים:** 429 ללקוחה שחרגה, 503 לשירות מלא; תמיד עם `Retry-After` בשניות שלמות.
- **לחץ חוזר:** חסימה כשהיצרן יכול לחכות, סירוב כשלא; ב־HTTP רק סירוב.
- **השלת עומס:** מקביליות ולא קצב; קודם מוסיפים ואז בודקים; מקום שמור לעדיפות גבוהה.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
L = lambda * W          W = S / (1 - rho)   (M/M/1)
ratelimit.NewBucket(rate, burst, now)  ->  Decision{Allowed, Remaining, RetryAfter}
rate.NewLimiter(rate.Limit(r), b).AllowN(t, n)        # golang.org/x/time/rate
redis.NewScript(src).Run(ctx, rdb, keys, args...)    # EVALSHA, falls back to EVAL on NOSCRIPT
redis.call('TIME')  HMGET/HSET  PEXPIRE               # inside the Lua script
429 + Retry-After: <seconds>      503 + Retry-After when shedding
RateLimit-Policy: "default";q=<burst>;w=<seconds>    RateLimit: "default";r=<remaining>
q.Put(ctx, v) blocks    q.TryPut(v) -> bounded.ErrFull
shed.New(shed.Options{Limit: 8, ReservedForHigh: 2, RetryAfter: time.Second})
docker compose -p ch16 -f data/compose.yaml up -d --wait   ...   down -v
my naive overshoot: ___ of 100 (limit 10)
my load run: no shed p99 ___  shed p99 ___  503 count ___
```

</div>

קישורים: [[15-timeouts-retries-breakers]] · [[17-consistency-replication-partitioning]]
