---
direction: rtl
tags: [senior-course, book]
chapter: 28
---
# אבטחה לשירותי backend

- **OWASP API Top 10 מ־2023:** BOLA ראשון, כי הוא נפוץ ובלתי נראה לכלים.
- **OIDC:** קוד עם PKCE למשתמשים, client credentials לשירותים; טוקנים קצרי חיים.
- **JWT:** אלגוריתם מהתצורה, `kid` עם סוג מפתח מתאים, `iss`, `aud`, `exp` חובה, מרווח קטן.
- **BOLA:** הדייר מהטוקן בחתימה של המאגר ובשאילתה; 404 זהה לחסר; בדיקה לכל נתיב.
- **SSRF:** בדיקת מחרוזת להודעה, `Control` של ה־dialer לביטחון; בלי proxy; הפניות נבדקות.
- **סודות:** משתנה סביבה, קובץ או מנהל סודות; Secret של Kubernetes הוא base64; החלפה עם חפיפה.
- **מפתחות API:** קידומת, מזהה ציבורי, 32 בתים; SHA-256 בלבד; השוואה בזמן קבוע.
- **לוגים:** `Secret`, `ReplaceAttr`, `Header`; maps לא מסוננים.
- **שרשרת אספקה:** SBOM עם Syft, Grype או Trivy, ו־govulncheck לצידם.
- **STRIDE:** שש שאלות על כל גבול אמון, ולכל שורה בדיקה ובעלים.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
OWASP API 2023: API1 BOLA, API2 AuthN, API3 property, API4 resources, API5 function, API6 flows,
                API7 SSRF, API8 misconfig, API9 inventory, API10 unsafe consumption
jwt.NewParser(WithValidMethods, WithIssuer, WithAudience, WithLeeway, WithExpirationRequired, WithIssuedAt, WithTimeFunc)
keyfunc: kid -> JWKS key; RSA methods need *rsa.PublicKey, ECDSA needs *ecdsa.PublicKey
401 + WWW-Authenticate: Bearer realm="...", error="invalid_token"
store.Get(ctx, tenantID, id)   SQL: WHERE tenant_id = $1 AND id = $2   -> 404 for foreign ids
net.Dialer{Control: policy.Control}   Transport{Proxy: nil}   CheckRedirect re-checks URL
blocked: 127/8 10/8 172.16/12 192.168/16 169.254/16 100.64/10 0/8 ::1 fc00::/7 fe80::/10 64:ff9b::/96
api key: <prefix>_<id hex>_<32 bytes base64url>; store sha256; subtle.ConstantTimeCompare
slog.HandlerOptions{ReplaceAttr: redact.ReplaceAttr}
docker run --rm -v "$PWD:/src:ro" anchore/syft:v1.54.0 scan dir:/src -o cyclonedx-json
docker run --rm -v "$PWD:/src:ro" anchore/grype:v0.120.0 dir:/src
STRIDE: Spoofing Tampering Repudiation Information-disclosure Denial-of-service Elevation
my STRIDE: top residual risk ___ owner ___
```

</div>

קישורים: [[27-code-review-static-analysis]] · [[29-system-design-method]]
