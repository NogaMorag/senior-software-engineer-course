---
direction: rtl
tags: [senior-course, book]
chapter: 33
---
# היתרון של Rust

- **הביקוש:** ‏Rust ב־2 מתוך 30 מודעות בכירות, כ־4% במדגם הרחב; ביטחון בינוני.
- **התחומים:** אבטחה, קריפטו, פלטפורמות תשתית ונתונים, מסחר אלגוריתמי, מערכות משובצות וכלי פיתוח.
- **שכר:** אין פרמיה מוכחת לשפה; התחום והמסלול משפיעים.
- **האסטרטגיה:** ‏Python בכירה ו־Go פותחות את השוק; ‏Rust היא הבידול.
- **החלטת שפה:** זנב, בטיחות בגבולות, מהירות הצוות, אקוסיסטם, ומה היה הופך את ההחלטה.
- **המחיר של שפה שנייה:** סקירה, תורנות, צנרת, גיוס.
- **קורות חיים:** השפה שלהם קודם, תוצאות ולא שפות, ‏Rust בסוגריים.
- **תרגום:** נתונים בתוך הנעילה מול לצידה; ‏`Option` מול תנאי; ‏`Arc` מול מצביע; הפלה מול `ctx.Done()`.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
demand       Rust 2/30 senior hand-read | 8/178 (~4%) keyword | Go 13/30 | SO 2025: Rust 14.5%, Go 17.4%
where        security (endpoint, network, data classification) | crypto | infra/data platforms | HFT | embedded/OS
choose Rust  tight p99 on a hot path (measure GC first) | untrusted input or FFI | slow-changing, owned component
choose Go    cloud-native ecosystem (client-go, controller-runtime, OTel) | team speed | many reviewers
2nd language reviewers | on-call | CI and supply chain | hiring
translate    Mutex<State> vs mu + fields | Option from checked_duration_since vs elapsed > 0
             Arc::clone from a locked map vs pointer after Unlock | drop Future vs return on ctx.Done()
             Result + ? vs if err != nil + %w; match vs errors.Is / errors.As | trait vs implicit interface
port         cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test
my sentence  For ___ I would pick ___ because ___; not for ___ because ___
```

</div>

קישורים: [[32-ai-assisted-engineering]] · [[34-senior-interview-loop]]
