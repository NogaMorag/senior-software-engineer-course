---
direction: rtl
lang: he
---

<div dir="rtl" lang="he">

# תרגיל פרק 33: דלי האסימונים, ב־Rust

**מטרה:** להעביר את דלי האסימונים מפרק 16 מ־Go ל־Rust: דלי אחד עם שעון מוזרק, ודלי לכל מפתח, עם אותן בדיקות
מקרה במקרה. ואז לכתוב עמוד אחד שמשווה בין שתי הגרסאות, ומנסח מתי היית בוחרת בכל אחת מהשפות. החלק של Rust
**אופציונלי**: אם אין לך עדיין Rust על המחשב החדש ואת לא רוצה להתקין, דלגי לצעד 7 וכתבי את עמוד ההשוואה מתוך
הקוד של פרק 16 ומהניסיון שלך.

**זמן משוער:** יום לימוד אחד, יום 92: כ־40 דקות בנייה (כ־25 להעברה ולבדיקות, כ־15 לעמוד ההשוואה).

**ידע מקדים:** פרק 33. מפרק 16, סעיף 16.2: דלי האסימונים ב־Go, בקבצים `bucket.go` ו־`local.go` של התרגיל שלו,
והבדיקות שלו. ‏Rust עצמה מהניסיון שלך: התרגיל לא מלמד אותה. אין תלויות חיצוניות, רק הספרייה הסטנדרטית.

## מה בתיקייה הזאת

| קובץ | תפקיד |
| --- | --- |
| `starter/Cargo.toml` | ה־crate בשם `tokenbucket`, ‏edition 2024, בלי תלויות |
| `starter/rust-toolchain.toml` | גרסת Rust המוצמדת, עם rustfmt ו־clippy; ‏rustup מתקין אותה לבד |
| `starter/src/lib.rs` | הטיפוסים מוכנים; `Bucket::allow_n` ו־`Local::allow` עם `todo!` |
| `starter/tests/bucket.rs` | ארבע בדיקות: שלוש מהבדיקות של Go, ובדיקה של שיתוף בין threads |
| `tests/bucket.rs` | הבדיקות הקנוניות; עותק זהה ב־`starter/` וב־`solution/` |
| `data/comparison-template.md` | התבנית לעמוד ההשוואה |
| `solution/src/lib.rs` | ההעברה המלאה |
| `solution/COMPARISON.md` | עמוד השוואה ממולא, לדוגמה |

## צעדים

**1. בדקי אם יש לך Rust.**

<div dir="ltr" align="left">

```bash
cargo --version
```

</div>

אם הפקודה לא נמצאה ואת רוצה לעשות את ההעברה, התקיני את rustup בתיקיית הבית, בלי לשנות קובצי הגדרות של המעטפת,
וטעני את הסביבה לחלון הנוכחי. אם לא, דלגי לצעד 7.

<div dir="ltr" align="left">

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- --no-modify-path -y --profile minimal
. "$HOME/.cargo/env"
cargo --version
```

</div>

כדי שזה יעבוד גם בחלונות חדשים, הוסיפי את השורה השנייה לסוף `~/.bashrc`. הגרסה המוצמדת של התרגיל, נכון לאוקטובר
2026, רשומה ב־`rust-toolchain.toml` ובנספח ג; ‏rustup יוריד אותה בפעם הראשונה שתריצי `cargo` בתיקייה.

**2. העתיקי את קוד ההתחלה, והריצי את הבדיקות.** הן נכשלות עם הודעות `todo!`:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course/ch33
cp -R "$COURSE/exercises/ch33-rust-edge/starter/." ~/senior-course/ch33/
cp "$COURSE/exercises/ch33-rust-edge/data/comparison-template.md" ~/senior-course/ch33/COMPARISON.md
cd ~/senior-course/ch33
grep -qx 'target/' ~/senior-course/.gitignore 2>/dev/null || echo 'target/' >> ~/senior-course/.gitignore
cargo test
```

</div>

**3. פתחי את גרסת Go לידה.** הקבצים `bucket.go` ו־`local.go` נמצאים בתיקייה של פרק 16:

<div dir="ltr" align="left">

```bash
less "$COURSE/exercises/ch16-ratelimit/solution/ratelimit/bucket.go"
```

</div>

**4. מלאי את `Bucket::allow_n`** בקובץ `src/lib.rs`, לפי ששת הצעדים שבהערה. ההערה מציינת את ההבדל היחיד שאין
לו מקבילה ישירה: `Instant` לא יכול ללכת אחורה, ולכן „שעון שחזר אחורה” נבדק עם `checked_duration_since`, שמחזירה
`None` במקרה כזה. הריצי את בדיקות הדלי בלבד:

<div dir="ltr" align="left">

```bash
cargo test bucket_table
cargo test clock_going_backwards_earns_nothing
cargo test shared_between_threads_never_over_allows
```

</div>

**5. מלאי את `Local::allow`.** אותו מבנה כמו ב־Go: נועלים את המפה רק כדי למצוא או ליצור את הדלי, משכפלים את
ה־`Arc`, משחררים את המפה, ורק אז קוראים לדלי.

<div dir="ltr" align="left">

```bash
cargo test
```

</div>

**6. מחקי את השורה `#![allow(dead_code)]`** מראש `src/lib.rs`, והריצי את שלושת הכלים. שלושתם צריכים לעבור בלי
הערה אחת:

<div dir="ltr" align="left">

```bash
cargo fmt --check
cargo clippy --all-targets -- -D warnings
cargo test
```

</div>

**7. מלאי את `COMPARISON.md`.** כל תא מתוך מה שעשית או מדדת. אם לא עשית את ההעברה, מלאי את עמודת Rust ממה שאת
יודעת, ורשמי „לא הועבר” בראש העמוד. ספירת שורות, אם עשית את ההעברה:

<div dir="ltr" align="left">

```bash
wc -l "$COURSE/exercises/ch16-ratelimit/solution/ratelimit/bucket.go" "$COURSE/exercises/ch16-ratelimit/solution/ratelimit/local.go" src/lib.rs
```

</div>

רק אחרי שסיימת, פתחי את `solution/COMPARISON.md` והשווי.

**8. שמרי את העבודה.**

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git add .gitignore ch33
git commit -m "ch33: token bucket ported to Rust, Rust vs Go comparison"
```

</div>

## פלט צפוי

בצעד 2, ארבע הבדיקות נכשלות על ה־`todo!`:

<div dir="ltr" align="left">

```text
test bucket_table ... FAILED
test clock_going_backwards_earns_nothing ... FAILED
test local_keys_are_independent ... FAILED
test shared_between_threads_never_over_allows ... FAILED
```

</div>

ובצעד 6, אחרי ההעברה (הסדר וזמן הריצה אצלך עשויים להיות אחרים):

<div dir="ltr" align="left">

```text
running 4 tests
test clock_going_backwards_earns_nothing ... ok
test bucket_table ... ok
test local_keys_are_independent ... ok
test shared_between_threads_never_over_allows ... ok

test result: ok. 4 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s
```

</div>

## רשימת בדיקה

- [ ] ארבע הבדיקות עוברות, כולל טבלת חמשת המקרים שהועברה מ־Go.
- [ ] השורה `#![allow(dead_code)]` נמחקה, והכלים של העיצוב ושל clippy שקטים.
- [ ] אין תלויות ב־`Cargo.toml`, והתיקייה `target/` לא נכנסה ל־Git.
- [ ] בעמוד ההשוואה יש לפחות מקרה אחד שבו היית בוחרת Rust ומקרה אחד שבו Go, כל אחד עם המחיר שלו.
- [ ] המשפט לראיון כתוב, בצורה „בשביל X הייתי בוחרת Y, כי...; לא בשביל Z, כי...”.
- [ ] העבודה שמורה ב־Git.

## טעויות נפוצות ותיקונים

- **`cargo` לא נמצא בחלון חדש.** ההתקנה לא שינתה את קובצי ההגדרות, בכוונה. הוסיפי את השורה שטוענת את
  `~/.cargo/env` לסוף `~/.bashrc`, או טעני אותה ידנית.
- **המקרה של בקשה גדולה מהדלי נכשל על `remaining`.** החזרת אפס במקום מה שיש בדלי. ב־Go, גם בקשה
  שלעולם לא תצליח מחזירה את מספר האסימונים שנשארו, אחרי המילוי.
- **המקרה `refill` נכשל עם יותר אסימונים מהצפוי.** שכחת לעדכן את `last` אחרי המילוי, ולכן אותו זמן נספר שוב
  בכל קריאה. עדכני אותו באותו תנאי שבו את ממלאת.
- **‏clippy מתלונן על שדות שלא נקראים.** השארת את `#![allow(dead_code)]` ואז מחקת אותו לפני שסיימת את שתי
  הפונקציות. סיימי את שתיהן, ואז מחקי.
- **הקומפיילר מסרב להחזיר הפניה לדלי מתוך המפה הנעולה.** ההפניה חיה רק כל עוד הנעילה מוחזקת. שכפלי את ה־`Arc`
  בתוך בלוק, ושחררי את הנעילה בסוף הבלוק.
- **‏`git status` מציג אלפי קבצים מתחת ל־`target/`.** השורה לא נוספה ל־`.gitignore`. הוסיפי אותה כמו בצעד 2.

## רמזים

- המתודה `as_secs_f64` היא המקבילה של `Seconds` ב־Go.
- המתודה `f64::min` היא המקבילה של `math.Min`.
- שרשור של `if let` עם תנאי נוסף מותר ב־edition 2024, ומחליף שני תנאים מקוננים.
- המרה עם `as` ממספר עשרוני לשלם חותכת לכיוון אפס, בדיוק כמו `int()` ב־Go.
- ‏`HashMap::entry` עם `or_insert_with` מוצא או יוצר בצעד אחד, ומחזיר הפניה שאפשר לשכפל ממנה את ה־`Arc`.

## הפתרון

ההעברה המלאה נמצאת ב־`solution/src/lib.rs`. היא עוקבת אחרי Go שורה בשורה: המילוי עצלני ונעשה רק כשמגיעה בקשה,
לפי השעון שהוזרק, ולכן הבדיקות לא ישנות. ההבדלים הם במקומות שבהם הטיפוסים של Rust עושים עבודה שב־Go עושה מוסכמה:
השדות נמצאים בתוך ה־`Mutex` ולא לצידו, ולכן אי אפשר לקרוא אותם בלי לנעול. השעון הוא פרמטר גנרי ולא פונקציה,
ושיתוף בין threads נבדק בזמן קומפילציה, כשב־Go הוא נבדק עם `-race` בזמן ריצה.

## מה התרגיל מדגים

- העברה של רכיב קטן בין השפות, עם אותן בדיקות, כדי להשוות על אותו קוד ולא על דעות.
- שעון מוזרק בשתי הצורות: פונקציה כשדה ב־Go, ‏trait גנרי ב־Rust.
- ‏`Mutex` שמחזיק את הנתונים, מול נעילה שיושבת לצד הנתונים ובדיקת מרוצים.
- ‏`Option` במקום בדיקה ידנית של זמן שלילי, ו־`Arc` בתוך מפה נעולה מול מצביע בתוך מפה.
- עמוד השוואה ומשפט לראיון: מתי Rust, מתי Go, ומה כל בחירה עולה לצוות.

</div>
