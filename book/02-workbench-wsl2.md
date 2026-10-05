---
direction: rtl
lang: he
title: "פרק 2: סביבת העבודה: WSL2, Docker, Go ו־uv"
chapter: 2
part: "חלק 0: לפני שמתחילים"
---

<div dir="rtl" lang="he">

# פרק 2: סביבת העבודה: WSL2, Docker, Go ו־uv

בסוף הפרק הזה יהיה לך מחשב שמריץ כל דבר בספר: מערכת הפעלה של שרתים בתוך Windows, מנוע קונטיינרים, שתי
שפות התכנות של הספר וניהול גרסאות, ובהמשך גם את הכלים של חלק ה. כל כלי מותקן באותה תבנית קבועה, כך שתמיד תדעי מה הוא, איפה הוא
יושב, איך מוודאים שהוא עובד, ואיך מסירים אותו בלי להשאיר לכלוך. בעבודה אמיתית, סביבה שאפשר להקים מחדש
בשעה אחת היא מה שמבדיל בין מפתחת שתלויה במזל לבין מפתחת שאפשר לסמוך עליה.

הפרק מחולק לשלושה ימי לימוד, בכוונה, כי יום התקנה ויום שימוש ראשון לא נפגשים:

| יום | סעיפים | מה עושים |
| --- | --- | --- |
| יום 0 | 2.1 עד 2.8 | מתקינים ומוודאים גרסאות. בלי שעון, בסוף השבוע שלפני ההתחלה. |
| יום 3 | 2.9, 2.11 ו־2.12 | מאגר הקורס, בדיקת העשן, ההתחייבות הראשונה, ועצירה וניקוי. |
| יום 57 | 2.10 | הכלים של חלק ה, יום לפני שמשתמשים בהם. |

על Mac עושים את אותם ימים עם סעיף 2.13 במקום ההוראות של Windows.

מה הורץ בפועל בזמן כתיבת הפרק, ומה לא:

| מה | איך נבדק |
| --- | --- |
| ההתקנה של WSL, וההגדרות של Docker Desktop ושל VS Code | לא הורצו על Windows; נכתבו מתוך התיעוד הרשמי |
| ההתקנות בתוך Ubuntu, `doctor.sh` ובדיקות העשן | הורצו בקונטיינר של Ubuntu 26.04; הפלטים הועתקו משם |
| אותן הורדות בגרסאות של macOS, וקונטיינר ה־hello-world | הורצו על Mac (סעיף 2.13) |

אם משהו ב־Windows נראה אצלך אחרת, התיעוד הרשמי הוא המקור, ונספח ב מרכז את התקלות הידועות.

## מה נלמד בפרק

- מה מתקינים, למה כל כלי נחוץ, ואיפה בדיוק הוא יושב על הדיסק.
- להתקין את WSL2 עם Ubuntu, ולעבוד בתוך מערכת הקבצים שלה.
- להתקין את Docker Desktop, ולבדוק את הדרישות שלו מראש.
- להגדיר את Git ומפתח SSH.
- לערוך ב־VS Code את הקבצים שבתוך המערכת.
- להתקין את Go ואת uv בתיקיית הבית, ולבדוק כל הורדה מול טביעת הקובץ שלה.
- להתקין בגרסה מוצמדת את כלי ה־Go שהספר משתמש בהם.
- להקים את מאגר הקורס, להריץ את `doctor.sh` ובדיקות עשן, ולבצע התחייבות ראשונה.
- לעצור, להפעיל מחדש, לנקות ולהסיר כל דבר שהתקנת.

## ידע מקדים

פרק 0 (המוסכמות של הספר). מחשב Windows 11 עם הרשאות מנהלת, חיבור לרשת, וחשבון GitHub. לא נדרש ידע
קודם ב־Linux: כל פקודה מוסברת.

## 2.1 מה מתקינים ולמה

| כלי | מה הוא עושה בקורס | איפה הוא יושב | מתי |
| --- | --- | --- | --- |
| WSL2 ו־Ubuntu | מערכת Linux אמיתית שבה רץ הכול | דיסק וירטואלי בתיקיית המשתמשת של Windows | יום 0 |
| Windows Terminal ו־bash | הטרמינל והמעטפת שבהם מקלידים כל פקודה | מובנה ב־Windows 11; bash בתוך Ubuntu | יום 0 |
| Docker Desktop | מריץ PostgreSQL, Redis, Redpanda ו־kind בקונטיינרים | אפליקציה ב־Windows ומכונת Linux קטנה משלה | יום 0 |
| Git ומפתח SSH | ניהול גרסאות, והזדהות מול GitHub בלי סיסמה | חבילת Ubuntu; המפתח בתיקייה `~/.ssh` | יום 0 |
| VS Code והרחבת WSL | העורך, שעובד ישירות על קבצי Linux | אפליקציה ב־Windows ושרת קטן בתוך Ubuntu | יום 0 |
| Go | השפה הראשית של הספר ושל פרויקט הגמר | התיקייה `~/.local/go` | יום 0 |
| כלי Go | ניתוח סטטי ובודקי קוד, מיגרציות, protobuf, סריקת חולשות, בדיקת workflows | התיקייה `~/.local/bin` | יום 0 |
| uv ו־Python 3.14 | פרויקטי Python, גרסאות וקבצי נעילה | התיקייה `~/.local/bin` ומטמון של uv | יום 0 |
| kind, kubectl, Helm, Terraform, k6 | הכלים של חלק ה | התיקייה `~/.local/bin` | יום 57 |

מה לא מתקינים, ולמה:

- **לא מנוע Docker נוסף מתוך `apt`.** הוא מתנגש ב־Docker Desktop, ושני מנועים זה מקור קבוע לבלבול.
- **לא Go מתוך `apt`.** הגרסה בחבילות של ההפצה מפגרת אחרי הגרסה שהספר מצמיד.
- **לא pip ולא Python מהאתר הרשמי.** uv מנהל את Python ואת התלויות, עם קובץ נעילה.
- **לא `sudo` לכלים של הקורס.** כל הכלים חוץ מחבילות הבסיס נכנסים לתיקיית הבית שלך: קל לשדרג, קל
  להסיר, ושום דבר לא נוגע במערכת.

לכל כלי בפרק יש אותו מבנה, כדי שתדעי מה לחפש: מה ולמה, איפה, גרסה (עם הפניה לנספח ג), פקודות התקנה, פלט
צפוי, בדיקה, הפעלה מחדש, הסרה, וטבלת תקלות קצרה. התקלות שחוצות כלים מרוכזות בסעיף 2.11, וכל הטבלאות
מרוכזות שוב בנספח ב.

> **מונח:** משתנה הסביבה `PATH` הוא רשימת תיקיות, מופרדות בנקודתיים, שבהן המעטפת מחפשת פקודה כשאת מקלידה
> את שמה. „התקנה” של כלי שורת פקודה היא בדרך כלל רק העתקת קובץ הרצה לתיקייה שנמצאת ב־`PATH`. זו הסיבה
> שאחרי התקנה צריך לפעמים לפתוח טרמינל חדש: המעטפת קוראת את הגדרות ה־`PATH` רק כשהיא עולה.

## 2.2 WSL2 ו־Ubuntu

| שאלה | תשובה |
| --- | --- |
| מה | שכבה שמריצה ליבת Linux אמיתית במכונה וירטואלית קלה, ומעליה הפצת Ubuntu |
| למה | הכלים בספר נבנו קודם ל־Linux, ושרתי הייצור שתעבדי מולם רצים על Linux |
| איפה | קובץ דיסק וירטואלי (`ext4.vhdx`) בתיקיית המשתמשת של Windows; הבית שלך ב־Linux הוא `/home/<שם>` |
| גרסה | נכון לאוקטובר 2026: WSL 3.0.1, והפצת ברירת המחדל היא Ubuntu 26.04 LTS (נספח ג) |

**לפני ההתקנה.** צריך שני דברים:

- **מערכת הפעלה:** Windows 11, או Windows 10 בגרסה 22H2.
- **וירטואליזציה מופעלת בהגדרות הקושחה.** כדי לבדוק, פתחי את מנהל המשימות ועברי ללשונית הביצועים. בחרי
  את המעבד, ומצאי את השורה `Virtualization`: צריך להופיע בה `Enabled`. אם כתוב `Disabled`, מפעילים אותה
  בהגדרות הקושחה של המחשב (השם משתנה בין יצרנים).

**התקנה.** זו אחת הפקודות הבודדות בספר שרצות ב־Windows עצמו, ב־PowerShell עם הרשאות מנהלת. פותחים אותו
מהתפריט של לחיצה ימנית על סמל הטרמינל:

<div dir="ltr" align="left">

```powershell
wsl --install
```

</div>

הפקודה מפעילה את רכיבי Windows הנדרשים ומורידה את Ubuntu. בסיומה צריך **להפעיל מחדש את Windows**. אחרי
ההפעלה מחדש נפתח חלון של Ubuntu ומבקש שם משתמשת וסיסמה. אלה של Linux, לא של Windows:

- **שם משתמשת:** קצר, באותיות קטנות באנגלית, בלי רווחים. הוא יופיע בכל נתיב.
- **סיסמה:** תשמש את `sudo`, הפקודה שמריצה פעולה כמנהלת. בזמן ההקלדה לא מופיע דבר על המסך; זה תקין.

אם החלון לא נפתח מעצמו, או שאת רוצה לראות אילו הפצות קיימות, אלה הפקודות, שוב ב־PowerShell:

<div dir="ltr" align="left">

```powershell
wsl --list --online
wsl --install -d Ubuntu
```

</div>

**בדיקה מצד Windows.** ב־PowerShell רגיל:

<div dir="ltr" align="left">

```powershell
wsl --version
wsl --list --verbose
```

</div>

<div dir="ltr" align="left">

```text
WSL version: 3.0.1.0
Kernel version: ...
...
  NAME      STATE           VERSION
* Ubuntu    Running         2
```

</div>

העמודה החשובה היא VERSION: היא חייבת להיות 2. המספרים בשורות הראשונות יכולים להיות גבוהים יותר אצלך.

**בדיקה מצד Ubuntu.** מכאן והלאה, כל הפקודות רצות בתוך Ubuntu. פתחי את Ubuntu והריצי:

<div dir="ltr" align="left">

```bash
head -2 /etc/os-release
uname -r
sudo apt update && sudo apt upgrade -y
```

</div>

<div dir="ltr" align="left">

```text
PRETTY_NAME="Ubuntu 26.04.1 LTS"
NAME="Ubuntu"
...-microsoft-standard-WSL2
```

</div>

השורה השלישית מעדכנת את רשימת החבילות ואת החבילות המותקנות. היא שואלת את סיסמת ה־Linux שבחרת, ובפעם
הראשונה עשויה לקחת כמה דקות.

**הכלל החשוב ביותר בפרק: עובדים בתוך מערכת הקבצים של Linux.** הכוננים של Windows מופיעים תחת `/mnt/c`, וזה
מפתה לעבוד שם. אל תעשי זאת. גישה מ־Linux לקבצים של Windows עוברת דרך שכבת תרגום, והיא
איטית בהרבה; גם הרשאות הקבצים וסופי השורות מתנהגים אחרת. התיעוד של Microsoft ממליץ לשמור את קבצי
הפרויקט במערכת הקבצים של Linux, כשעובדים עם הכלים שלה. כל מה שבספר חי תחת `~`, תיקיית הבית שלך ב־Linux.

כדי לראות את הקבצים האלה בסייר הקבצים של Windows, הריצי את הפקודה הבאה מתוך Ubuntu. היא פותחת את התיקייה
הנוכחית בסייר:

<div dir="ltr" align="left">

```bash
explorer.exe .
```

</div>

**הפעלה מחדש.** הפקודה הראשונה למטה עוצרת את כל ההפצות ואת המכונה הווירטואלית; הטרמינל הבא שתפתחי יעלה
אותן מחדש. השנייה עוצרת רק את Ubuntu. שתיהן ב־PowerShell:

<div dir="ltr" align="left">

```powershell
wsl --shutdown
wsl --terminate Ubuntu
```

</div>

**גיבוי והסרה.** הפקודה `--unregister` מוחקת את ההפצה וכל מה שבתוכה, בלי דרך חזרה. לכן מגבים קודם לקובץ
אחד, שאפשר לשחזר ממנו בפקודת הייבוא (`--import`). ב־PowerShell:

<div dir="ltr" align="left">

```powershell
wsl --export Ubuntu D:\backup\ubuntu.tar
wsl --unregister Ubuntu
```

</div>

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| שגיאה `0x80370102` | וירטואליזציה כבויה בקושחה | להפעיל אותה בהגדרות ה־UEFI ולהפעיל מחדש |
| שגיאה `0x8007019e` | רכיב ה־WSL של Windows לא הופעל | להריץ שוב את פקודת ההתקנה כמנהלת, ולהפעיל מחדש |
| ההפצה לא מופיעה ברשימה | ההורדה לא הושלמה | להתקין במפורש עם `-d Ubuntu` |
| שכחת את סיסמת ה־Linux | | ב־PowerShell: `wsl -u root`, ואז בתוך Ubuntu `passwd <שם>` |

## 2.3 הטרמינל ו־bash

הטרמינל של Windows מותקן מראש. בהגדרות שלו (החץ הקטן ליד הלשוניות) עשי שני שינויים:

| איפה בהגדרות | מה לשנות |
| --- | --- |
| `Startup`, השדה `Default profile` | לבחור Ubuntu |
| הפרופיל של Ubuntu, השדה `Command line` | השורה שמתחת |

השינוי השני חשוב: בלעדיו הטרמינל נוטה להיפתח בתיקייה של Windows, תחת `/mnt/c`, בדיוק המקום שבו לא עובדים.

<div dir="ltr" align="left">

```text
wsl.exe -d Ubuntu --cd ~
```

</div>

פתחי לשונית חדשה. הפקודה `pwd` מדפיסה את התיקייה הנוכחית, וצריכה להדפיס `/home/<שם>`. בספר הפקודות
מופיעות בלי התחילית שאת רואה במסך (שם, מחשב ונתיב), כדי שתוכלי להעתיק אותן כמו שהן.

המעטפת (shell) היא התוכנית שקוראת את מה שאת מקלידה ומריצה אותו. ב־Ubuntu זו bash. הפקודות שיחזרו בכל
הספר:

| פקודה | מה היא עושה |
| --- | --- |
| `pwd` | מדפיסה את התיקייה הנוכחית; הבדיקה הראשונה כשמשהו „לא נמצא” |
| `ls -la` | רשימת הקבצים, כולל קבצים נסתרים שמתחילים בנקודה |
| `cd ~` | חזרה לתיקיית הבית; `cd -` חוזרת לתיקייה הקודמת |
| `mkdir -p a/b` | יצירת תיקייה, כולל תיקיות הביניים |
| `cat file` ו־`less file` | הדפסת קובץ, או דפדוף בו (יציאה עם q) |
| `command -v go` | מאיזה קובץ תרוץ הפקודה; ריק אם היא לא נמצאת |
| `echo "$PATH"` | התיקיות שבהן המעטפת מחפשת פקודות |
| `source ~/.bashrc` | קריאה מחדש של קובץ ההגדרות של bash, בלי לפתוח טרמינל |
| `Ctrl+C` ו־`Ctrl+R` | עצירת תוכנית שרצה; חיפוש בהיסטוריית הפקודות |

**חבילות בסיס.** אלה החבילות היחידות שמותקנות עם `sudo`, פעם אחת:

<div dir="ltr" align="left">

```bash
sudo apt install -y build-essential ca-certificates curl jq unzip
```

</div>

| חבילה | למה צריך אותה |
| --- | --- |
| `build-essential` | מהדר C; גלאי המרוצים של Go (פרק 5) צריך אותו |
| `ca-certificates` | התעודות שמאפשרות הורדות מאובטחות |
| `curl` | הורדת קבצים מהטרמינל |
| `jq` | קריאת JSON מהטרמינל; נשתמש בו לבדיקת ההורדה של Go |
| `unzip` | פתיחת קבצי zip; Terraform מגיע כך (סעיף 2.10) |

**התיקייה של הכלים, ב־`PATH`.** כל הכלים של הקורס נכנסים לשתי תיקיות בבית שלך:

- **ל־Go עצמה:** `~/.local/go/bin`
- **לכל השאר:** `~/.local/bin`

הפקודות האלה יוצרות את השנייה ומוסיפות את שתיהן לתחילת ה־`PATH`, בקובץ ההגדרות של bash:

<div dir="ltr" align="left">

```bash
mkdir -p ~/.local/bin
cat >> ~/.bashrc <<'EOF'

# senior-course (chapter 2): user-space tools first
export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH"
EOF
source ~/.bashrc
echo "$PATH" | tr ':' '\n' | head -2
```

</div>

<div dir="ltr" align="left">

```text
/home/noga/.local/go/bin
/home/noga/.local/bin
```

</div>

הסימן `>>` מוסיף לסוף הקובץ ולא דורס אותו, וזה חשוב: בקובץ הזה יש הגדרות של Ubuntu. תיקייה שמופיעה
מוקדם ב־`PATH` מנצחת, ולכן גרסה שהתקנת בבית תנצח גרסה ישנה של המערכת. הריצי את בלוק הפקודות פעם אחת
בלבד; הרצה שנייה תוסיף שורה כפולה (לא מזיק, אבל מבלבל).

**עברית בטרמינל.** הטרמינל מציג אותיות עבריות, אבל ייתכן שמשפט שלם יופיע בסדר הפוך. זו בעיית תצוגה בלבד:
הקובץ עצמו תקין, ו־VS Code יציג אותו נכון. הקידוד צריך להיות UTF-8. כדי לבדוק, הריצי `locale`: השורה
הראשונה בפלט צריכה להסתיים ב־`UTF-8`.

**הפעלה מחדש והסרה.** הפקודה `exec bash` מחליפה את המעטפת הנוכחית במעטפת טרייה, כמו לשונית חדשה. אין מה
להסיר: הטרמינל מובנה. כדי לבטל את שינוי ה־`PATH`, מוחקים את שלוש השורות שהוספת בסוף `~/.bashrc`.

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| הטרמינל נפתח ב־`/mnt/c/Users/...` | תיקיית ההתחלה של הפרופיל | שדה Command line כמו למעלה |
| `command not found` מיד אחרי התקנה | המעטפת לא קראה מחדש את ה־`PATH` | `source ~/.bashrc`, או לשונית חדשה |
| `Could not get lock /var/lib/dpkg/lock-frontend` | עדכון אוטומטי רץ ברקע | לחכות דקה ולנסות שוב; לא למחוק את קובץ הנעילה |

## 2.4 Docker Desktop עם WSL2

| שאלה | תשובה |
| --- | --- |
| מה | אפליקציה של Windows שמריצה מנוע Docker במכונה וירטואלית משלה, ומחברת אליו את הפקודה `docker` |
| למה | כל שירות שהספר נשען עליו (PostgreSQL, Redis, Redpanda, אשכול kind) רץ בקונטיינר |
| איפה | האפליקציה ב־Windows; הקונטיינרים והאימג'ים בדיסק הווירטואלי של הפצת ה־WSL שלה |
| גרסה | נכון לאוקטובר 2026: Docker Desktop 4.93.0 עם Engine 29.8.1 (נספח ג) |
| עלות | חינם לשימוש אישי, לפי תנאי השימוש של Docker |

**הדרישות, לפני שמורידים.** כך מופיעות בעמוד ההתקנה הרשמי של Docker, כפי שנקרא ב־5 באוקטובר 2026:

| דרישה | פירוט |
| --- | --- |
| WSL | גרסה 2.1.5 ומעלה (בדיקה: `wsl --version` ב־PowerShell) |
| Windows 11 | גרסה 23H2 (build 22631) ומעלה |
| Windows 10 | גרסה 22H2 (build 19045) |
| מהדורה | העמוד מונה Pro, Enterprise ו־Education; אם אצלך Home, קראי את העמוד העדכני לפני שממשיכים |
| שירות Windows | השירות LanmanServer מופעל, ומוגדר להפעלה אוטומטית |
| חומרה | מעבד עם SLAT, לפחות 8 GB זיכרון, וירטואליזציה מופעלת |
| הרשאות | ההתקנה הרגילה, למשתמשת אחת, לא דורשת הרשאות מנהלת |
| מעבד ARM | נכון לאוקטובר 2026 הגרסה ל־Windows על ARM היא גישה מוקדמת (Early Access) בלבד |

את גרסת ה־Windows מציגה הפקודה `winver`, מתפריט התחל. את מצב השירות בודקים ב־PowerShell. אם
הוא לא רץ, או לא מוגדר אוטומטי, שתי הפקודות האחרונות מתקנות, ב־PowerShell של מנהלת:

<div dir="ltr" align="left">

```powershell
Get-Service LanmanServer | Select-Object Status, StartType
Set-Service LanmanServer -StartupType Automatic
Start-Service LanmanServer
```

</div>

**התקנה.** הורידי את המתקין מעמוד ההתקנה הרשמי (`docs.docker.com/desktop/setup/install/windows-install/`)
והריצי אותו. השאירי מסומנת את האפשרות להשתמש ב־WSL 2 במקום Hyper-V. אפשרות שקולה, ב־PowerShell:

<div dir="ltr" align="left">

```powershell
winget install -e --id Docker.DockerDesktop
```

</div>

אחרי ההתקנה הפעילי את Docker Desktop ואשרי את תנאי השימוש. בהגדרות שלו (גלגל השיניים):

| לשונית | מה לעשות |
| --- | --- |
| `General` | לוודא שהאפשרות `Use the WSL 2 based engine` מסומנת |
| `Resources`, ואז `WSL integration` | להדליק את השילוב עם הפצת ברירת המחדל ואת המתג של Ubuntu, וללחוץ `Apply & restart` |

**בדיקה, בתוך Ubuntu.** פתחי לשונית חדשה (השילוב מוסיף את הפקודה `docker` רק למעטפות חדשות):

<div dir="ltr" align="left">

```bash
docker version
docker compose version
docker info --format '{{.OperatingSystem}}'
```

</div>

<div dir="ltr" align="left">

```text
Client:
 Version:           29.8.1
 ...
Server: Docker Desktop 4.93.0 (...)
 Engine:
  Version:          29.8.1
 ...
Docker Compose version v...
Docker Desktop
```

</div>

מה שחשוב: יש גם חלק Client וגם חלק Server. חלק Server מוכיח שהפקודה ב־Ubuntu מדברת עם המנוע. את הקונטיינר
הראשון נריץ רק ביום 3, בסעיף 2.9.

**הפעלה מחדש.** בתפריט של הסמל שלו, באזור ההודעות, בוחרים Restart. אם המנוע לא מגיב: יוצאים מ־Docker
Desktop, מריצים את הפקודה שמתחת ב־PowerShell, ומפעילים אותו שוב. כשלא צריך אותו, אפשר לצאת ממנו לגמרי
ולפנות זיכרון.

<div dir="ltr" align="left">

```powershell
wsl --shutdown
```

</div>

**הסרה.** קודם מוחקים את הנתונים שלו (סעיף 2.12 מסביר מה נמחק), ואז מסירים את האפליקציה מהגדרות Windows,
בעמוד האפליקציות המותקנות. ההסרה מוחקת גם את הפצת ה־WSL הפנימית שלו.

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| `docker: command not found` ב־Ubuntu | השילוב עם Ubuntu כבוי, או לשונית ישנה | להדליק את המתג, ולפתוח לשונית חדשה |
| `Cannot connect to the Docker daemon` | Docker Desktop לא רץ, או עדיין עולה | להפעיל אותו ולחכות שיופיע שהמנוע רץ |
| `permission denied` על `docker.sock` | המעטפת נפתחה לפני שהשילוב הופעל | לסגור את הטרמינל, `wsl --shutdown`, לפתוח מחדש |
| המתקין מתלונן על גרסת WSL | WSL ישן מ־2.1.5 | `wsl --update` ב־PowerShell |
| שתי גרסאות `docker` מתנגשות | הותקן בעבר `docker.io` מ־`apt` | `sudo apt remove docker.io`, ולפתוח לשונית חדשה |

## 2.5 Git ומפתחות SSH

| שאלה | תשובה |
| --- | --- |
| מה | Git לניהול גרסאות; זוג מפתחות SSH להזדהות מול GitHub בלי סיסמה |
| למה | כל עבודה בספר נשמרת במאגר שלך, ופרויקט הגמר יפורסם ב־GitHub |
| איפה | Git מחבילות Ubuntu; ההגדרות ב־`~/.gitconfig`; המפתחות ב־`~/.ssh` |
| גרסה | הגרסה של ההפצה; `doctor.sh` דורש לפחות 2.28 |

<div dir="ltr" align="left">

```bash
sudo apt install -y git
git --version
```

</div>

<div dir="ltr" align="left">

```text
git version 2.53.0
```

</div>

זו הגרסה שהודפסה בקונטיינר Ubuntu 26.04 שבו נבדק הפרק; כל גרסה מ־2.28 ומעלה מספיקה. עכשיו הזהות שלך
והרגלים טובים. השתמשי בשם ובמייל האישיים, לא של מקום העבודה: המאגר הזה ילווה אותך לתיק העבודות.

<div dir="ltr" align="left">

```bash
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
git config --global init.defaultBranch main
git config --global core.autocrlf input
git config --global --list
```

</div>

| הגדרה | למה |
| --- | --- |
| `user.name` ו־`user.email` | מופיעים בכל התחייבות |
| `init.defaultBranch main` | כל מאגר חדש מתחיל בענף בשם `main` |
| `core.autocrlf input` | אם נכנס קובץ עם סופי שורה של Windows, Git ישמור אותו עם סופי שורה של Linux |

**מפתח SSH.** זוג מפתחות הוא שני קבצים: מפתח פרטי שלא יוצא מהמחשב לעולם, ומפתח ציבורי (הסיומת `.pub`)
שנותנים ל־GitHub. כש־GitHub רואה חיבור, הוא בודק שלצד השני יש את המפתח הפרטי שמתאים.

<div dir="ltr" align="left">

```bash
ssh-keygen -t ed25519 -C "you@example.com"
cat ~/.ssh/id_ed25519.pub
```

</div>

לשאלה על מיקום הקובץ לוחצים Enter (ברירת המחדל). לשאלה על סיסמת מפתח (passphrase) מומלץ לבחור אחת: היא
מגינה על המפתח אם הקובץ דולף. הפקודה השנייה מדפיסה שורה אחת שמתחילה ב־`ssh-ed25519`. העתיקי אותה, ואז:

1. ב־GitHub, פתחי את ההגדרות של החשבון.
2. בעמוד `SSH and GPG keys`, לחצי על `New SSH key`.
3. הדביקי את השורה, תני למפתח שם (למשל שם המחשב), ושמרי.

כדי לא להקליד את סיסמת המפתח בכל פעם, נשתמש בכלי `keychain`, שמחזיק את המפתח פתוח בזיכרון בין לשוניות:

<div dir="ltr" align="left">

```bash
sudo apt install -y keychain
echo 'eval "$(keychain --eval --quiet id_ed25519)"' >> ~/.bashrc
source ~/.bashrc
ssh -T git@github.com
```

</div>

בחיבור הראשון מוצגת טביעת אצבע של השרת, עם שאלה אם להמשיך. השווי אותה לטביעות שבתיעוד של GitHub, בעמוד
שכתובתו מתחת, ורק אז הקלידי `yes`:

<div dir="ltr" align="left">

```text
https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints
```

</div>

התשובה הצפויה:

<div dir="ltr" align="left">

```text
Hi <your-user>! You've successfully authenticated, but GitHub does not provide shell access.
```

</div>

קוד היציאה של הפקודה הזאת הוא 1 גם כשהכול תקין, כי GitHub לא נותן מעטפת. המשפט הוא ההוכחה.

**החלפה והסרה.** אם מפתח דלף או שאת מחליפה מחשב: מוחקים אותו בעמוד המפתחות ב־GitHub (זה מה שמבטל אותו
באמת), מוחקים את שני הקבצים, ויוצרים זוג חדש. את Git עצמו מסירה הפקודה הראשונה:

<div dir="ltr" align="left">

```bash
sudo apt remove git
rm ~/.ssh/id_ed25519 ~/.ssh/id_ed25519.pub
```

</div>

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| `Permission denied (publickey)` | המפתח לא נוסף ל־GitHub, או לא טעון | לבדוק את עמוד המפתחות; `ssh-add -l` מראה מה טעון |
| `UNPROTECTED PRIVATE KEY FILE` | הרשאות רחבות מדי על המפתח | `chmod 700 ~/.ssh` ו־`chmod 600 ~/.ssh/id_ed25519` |
| כל הקבצים נראים „שונו” ב־`git status` | הקבצים נערכו מצד Windows עם סופי שורה אחרים | לעבוד רק דרך WSL; `core.autocrlf input` |

## 2.6 VS Code עם הרחבת WSL

| שאלה | תשובה |
| --- | --- |
| מה | העורך רץ ב־Windows, ושרת קטן שלו רץ בתוך Ubuntu ועושה את העבודה על הקבצים |
| למה | עריכה, טרמינל, ניפוי שגיאות והשלמת קוד, כולם רואים את סביבת ה־Linux ולא את Windows |
| איפה | האפליקציה ב־Windows; השרת שלה ב־`~/.vscode-server` בתוך Ubuntu |
| גרסה | הגרסה הנוכחית; `doctor.sh` בודק רק שהפקודה קיימת |

**התקנה, ב־Windows.** הורידי את המתקין מ־`code.visualstudio.com` והשאירי מסומנת את האפשרות להוסיף ל־PATH.
אפשרות שקולה ב־PowerShell היא השורה שמתחת. אחרי ההתקנה, בלשונית ההרחבות של העורך, חפשי את
ההרחבה `ms-vscode-remote.remote-wsl` והתקיני אותה.

<div dir="ltr" align="left">

```powershell
winget install -e --id Microsoft.VisualStudioCode
```

</div>

**שימוש ראשון, מתוך Ubuntu.** פתחי לשונית חדשה (ההתקנה שינתה את ה־PATH של Windows) והריצי:

<div dir="ltr" align="left">

```bash
cd ~
code .
```

</div>

בפעם הראשונה מופיעה הודעה על התקנת השרת, ואחרי כמה שניות נפתח חלון. בפינה השמאלית התחתונה צריך להופיע
`WSL: Ubuntu`. זה הסימן שהחלון עובד על קבצי Linux. ההרחבות מותקנות בנפרד לכל צד, ולכן מתקינים את אלה של
הקורס מתוך Ubuntu:

<div dir="ltr" align="left">

```bash
code --install-extension golang.go
code --install-extension ms-python.python
code --install-extension charliermarsh.ruff
code --version
```

</div>

הפקודה האחרונה מדפיסה שלוש שורות: הגרסה, מזהה הבנייה והארכיטקטורה. כשתפתחי לראשונה קובץ Go, אחרי סעיף
2.7, ההרחבה תציע להתקין כלי עזר (ביניהם `gopls`); אשרי.

**הפעלה מחדש.** מלוח הפקודות (Ctrl+Shift+P) מריצים „Developer: Reload Window”. אם השרת נתקע, סוגרים את העורך,
מוחקים את התיקייה `~/.vscode-server`, ופותחים שוב: השרת יותקן מחדש.

**הסרה.** מסירים את האפליקציה בהגדרות Windows, ומוחקים את השרת בתוך Ubuntu:

<div dir="ltr" align="left">

```bash
rm -rf ~/.vscode-server
```

</div>

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| `code: command not found` ב־Ubuntu | הלשונית נפתחה לפני ההתקנה | לשונית חדשה; אם לא עוזר, `wsl --shutdown` |
| אין `WSL: Ubuntu` בפינה, והנתיבים הם של Windows | התיקייה נפתחה מצד Windows | לסגור, ולפתוח עם `code .` מתוך Ubuntu |
| ההרחבה של Go מתלוננת שאין Go | Go עוד לא הותקן, או שהחלון נפתח לפניו | להשלים את סעיף 2.7 ולטעון מחדש את החלון |

## 2.7 Go

| שאלה | תשובה |
| --- | --- |
| מה | המהדר, הכלים (`go build`, `go test`, `gofmt`, `go vet`) והספרייה הסטנדרטית |
| למה | השפה הראשית של הספר ושל Hookline |
| איפה | שרשרת הכלים ב־`~/.local/go`; מטמון המודולים ב־`~/go`; כלים שתתקיני ב־`~/.local/bin` |
| גרסה | נכון לאוקטובר 2026: Go 1.27.1 (נספח ג) |

ההתקנה היא קובץ דחוס רשמי שנפתח לתוך תיקיית הבית, בלי `sudo`. לפני הפתיחה בודקים שהקובץ שהורד הוא בדיוק
הקובץ שפורסם. האתר של Go מפרסם טביעת SHA-256 לכל קובץ. הטביעה (checksum) היא מחרוזת שמחושבת מכל בית בקובץ. הפקודה `sha256sum`
מחשבת אותה מחדש ומשווה.

<div dir="ltr" align="left">

```bash
ARCH=$(dpkg --print-architecture)
GO_TGZ="go1.27.1.linux-${ARCH}.tar.gz"
cd /tmp
curl -fsSLO "https://go.dev/dl/${GO_TGZ}"
curl -fsSL "https://go.dev/dl/?mode=json&include=all" \
  | jq -r --arg f "$GO_TGZ" '.[].files[] | select(.filename == $f) | "\(.sha256)  \(.filename)"' \
  | sha256sum -c -
rm -rf "$HOME/.local/go"
tar -C "$HOME/.local" -xzf "$GO_TGZ"
go env -w GOBIN="$HOME/.local/bin"
go version
go env GOROOT GOPATH GOBIN
```

</div>

<div dir="ltr" align="left">

```text
go1.27.1.linux-amd64.tar.gz: OK
go version go1.27.1 linux/amd64
/home/noga/.local/go
/home/noga/go
/home/noga/.local/bin
```

</div>

| פקודה | מה היא עשתה |
| --- | --- |
| `dpkg --print-architecture` | הארכיטקטורה של המחשב: `amd64` ברוב המחשבים, `arm64` במחשבי ARM |
| `curl -fsSLO ...` | הורדת הקובץ לתיקייה הנוכחית, עם כישלון ברור אם השרת מחזיר שגיאה |
| `curl ... \| jq ... \| sha256sum -c -` | הטביעה הרשמית מתוך רשימת ההורדות, והשוואה לקובץ; חייב להודפס `OK` |
| `rm -rf` ואז `tar -C ... -xzf` | מחיקת גרסה קודמת, ופתיחת הקובץ לתוך `~/.local/go` |
| `go env -w GOBIN=...` | כלים שתתקיני עם `go install` ייכנסו ל־`~/.local/bin`, שכבר ב־`PATH` |

שימי לב למחיקה שלפני הפתיחה. התיעוד של Go מזהיר במפורש לא לפתוח גרסה חדשה מעל תיקייה קיימת, כי קבצים
ישנים שנשארים יוצרים התקנה שבורה. השדרוג הבא שלך יהיה אותן פקודות עם מספר גרסה אחר.

### 2.7.1 כלי Go שהספר משתמש בהם

כמה פרקים צריכים כלים שנכתבו ב־Go ומותקנים מקוד המקור. הפקודה `go install` מורידה מודול בגרסה מדויקת, בונה
אותו, ושמה את קובץ ההרצה ב־`GOBIN`, שהגדרנו ל־`~/.local/bin`. כך הכלים כבר ב־`PATH`, בלי שורה נוספת. מתקינים
אותם היום, ביום 0, כי פרק 9 משתמש ב־buf כבר ביום 21:

| כלי | למה | פרק |
| --- | --- | --- |
| staticcheck | ניתוח סטטי (static analysis) ל־Go, בכל תרגיל | 3 |
| buf, protoc-gen-go, protoc-gen-go-grpc | יצירת קוד מקובצי protobuf | 9 |
| goose | מיגרציות לסכמה של PostgreSQL | 11 |
| actionlint | בדיקת קובצי workflow של GitHub Actions | 23 |
| govulncheck | סריקת חולשות ידועות בתלויות | 27 |
| golangci-lint | מריץ עשרות בודקי קוד בבת אחת; מותקן כקובץ מוכן, לא דרך `go install` | 27 |

<div dir="ltr" align="left">

```bash
go install honnef.co/go/tools/cmd/staticcheck@2026.2.1
go install github.com/pressly/goose/v3/cmd/goose@v3.28.0
go install github.com/bufbuild/buf/cmd/buf@v1.73.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
staticcheck -version
buf --version
govulncheck -version
```

</div>

<div dir="ltr" align="left">

```text
staticcheck 2026.2.1 (0.8.1)
1.73.0
Go: go1.27.1
Scanner: govulncheck@v1.8.0
...
```

</div>

כל פקודה לוקחת כמה עשרות שניות, כי היא מורידה ובונה. הגרסאות מוצמדות (נספח ג): `@latest` היה נותן כלי אחר בכל
התקנה. שימי לב גם לפלט של govulncheck: השורה הראשונה היא גרסת Go, וגרסת הכלי מופיעה רק בשורה `Scanner`. הסקריפט
`doctor.sh` יודע לקרוא אותה משם.

**golangci-lint.** הפרויקט ממליץ להתקין אותו מקובץ מוכן ולא לבנות מהמקור, ולכן הוא מגיע כמו הכלים של סעיף
2.10: מורידים את הקובץ הדחוס ואת קובץ הטביעות מאותו פרסום, בודקים, ורק אז מתקינים. הפקודה `install -m 0755`
מעתיקה את הקובץ ל־`~/.local/bin` ומסמנת אותו כקובץ הרצה:

<div dir="ltr" align="left">

```bash
ARCH=$(dpkg --print-architecture)
GL=golangci-lint-2.14.0-linux-${ARCH}
cd /tmp
curl -fsSLO "https://github.com/golangci/golangci-lint/releases/download/v2.14.0/${GL}.tar.gz"
curl -fsSLO "https://github.com/golangci/golangci-lint/releases/download/v2.14.0/golangci-lint-2.14.0-checksums.txt"
sha256sum -c --ignore-missing golangci-lint-2.14.0-checksums.txt
tar -xzf "${GL}.tar.gz"
install -m 0755 "${GL}/golangci-lint" ~/.local/bin/golangci-lint
rm -rf "${GL}" "${GL}.tar.gz" golangci-lint-2.14.0-checksums.txt
golangci-lint --version
```

</div>

<div dir="ltr" align="left">

```text
golangci-lint-2.14.0-linux-amd64.tar.gz: OK
golangci-lint has version 2.14.0 built with go1.27.0 from 114493f9 on 2026-09-24T11:07:15Z
```

</div>

קובץ הטביעות מכיל שורה לכל קובץ בפרסום, והדגל `--ignore-missing` בודק רק את מה שהורד. הפלט הועתק מהרצה
במחשב שעליו נכתב הספר: בדיקת הטביעה על הקובץ של `linux-amd64`, ושורת הגרסה מהקובץ המקביל של macOS.

כלים אחרים שהספר משתמש בהם, כמו סורקי אימג'ים ובודק ה־OpenAPI, רצים כקונטיינרים בתוך הפרק שצריך אותם, ואין מה
להתקין מראש.

> **ב־Mac:** אין את הפקודה `dpkg`, ולכן קובעים את הארכיטקטורה ידנית; סעיף 2.13 נותן את השורה המקבילה.

> **איך Rust עושה את זה:** ב־Rust היית מתקינה עם rustup, שמנהל כמה שרשראות כלים ומעדכן אותן. ב־Go שרשרת
> הכלים היא תיקייה אחת שמחליפים בשלמותה, והקובץ `go.mod` של כל פרויקט מצהיר על הגרסה המינימלית. כשקובץ
> כזה דורש גרסה חדשה ממה שמותקן, Go יודע להוריד אותה בעצמו (ההגדרה `GOTOOLCHAIN`). התרגילים בספר מצהירים
> על `go 1.27`, ולכן אצלך שום דבר לא יורד מעצמו.

**הפעלה מחדש.** אין שירות לעצור. אם משהו מוזר, ניקוי מטמון הבנייה הוא הצעד הראשון:

<div dir="ltr" align="left">

```bash
go clean -cache
```

</div>

**הסרה.** קבצי מטמון המודולים מוגנים מכתיבה, ולכן `rm` רגיל נכשל עליהם. מנקים קודם עם Go עצמו, מבטלים את
ההגדרה שכתבת, ורק אז מוחקים את Go ואת הכלים שהותקנו דרכו. אחר כך מוחקים את השורה של `~/.local/go/bin`
מ־`~/.bashrc`:

<div dir="ltr" align="left">

```bash
go clean -modcache
go env -u GOBIN
rm -rf ~/.local/go ~/go
cd ~/.local/bin && rm staticcheck goose buf protoc-gen-go protoc-gen-go-grpc govulncheck actionlint
rm -rf ~/.local/bin/golangci-lint ~/.cache/golangci-lint
```

</div>

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| `go: command not found` | ה־`PATH` לא נטען | `source ~/.bashrc`; לבדוק שהשורה מסעיף 2.3 קיימת |
| `go version` מדפיס גרסה ישנה | Go מ־`apt` מוקדם ב־`PATH` | `command -v go`; אם `/usr/bin/go`, אז `sudo apt remove golang-go` |
| `sha256sum: WARNING: 1 computed checksum did NOT match` | הורדה פגומה או חלקית | למחוק ולהוריד שוב; לא לפתוח קובץ שנכשל בבדיקה |
| `cgo: C compiler "gcc" not found` | חסר מהדר C | חבילות הבסיס מסעיף 2.3 |

## 2.8 uv ו־Python

| שאלה | תשובה |
| --- | --- |
| מה | uv: מנהל גרסאות Python, סביבות וירטואליות ותלויות, עם קובץ נעילה |
| למה | Python היא השפה השנייה של הספר (פרקים 7 ו־8, והמקבל המדומה של Hookline) |
| איפה: הכלי | `uv` ו־`uvx` בתיקייה `~/.local/bin` |
| איפה: גרסאות Python | `~/.local/share/uv/python` |
| איפה: המטמון | `~/.cache/uv` |
| גרסה | נכון לאוקטובר 2026: uv 0.12.23 ו־Python 3.14 (נספח ג) |

המתקין הרשמי של uv הוא סקריפט קצר. הכתובת שמתחת כוללת את מספר הגרסה, ולכן תקבלי בדיוק את הגרסה שהספר
נבדק איתה. המתקין מזהה שהתיקייה `~/.local/bin` כבר ב־`PATH` (סעיף 2.3), ואם הוא מציע להוסיף שורה לקובץ
ההגדרות, זה לא מזיק.

<div dir="ltr" align="left">

```bash
curl -LsSf https://astral.sh/uv/0.12.23/install.sh | sh
uv --version
uv python install 3.14
uv python find 3.14
```

</div>

<div dir="ltr" align="left">

```text
downloading uv 0.12.23 x86_64-unknown-linux-gnu
installing to /home/noga/.local/bin
  uv
  uvx
everything's installed!
uv 0.12.23 (x86_64-unknown-linux-gnu)
Installed Python 3.14.8 in 1.76s
 + cpython-3.14.8-linux-x86_64-gnu (python3.14)
/home/noga/.local/share/uv/python/cpython-3.14-linux-x86_64-gnu/bin/python3.14
```

</div>

הזמן, וגרסת התיקון של Python (הספרה השלישית), יכולים להיות שונים אצלך. למערכת יש Python משלה, בשם
`python3`: היא שייכת למערכת ההפעלה, ואנחנו לא נוגעות בה. כל פרויקט בספר מצהיר על Python 3.14 בקובץ
`pyproject.toml`, ו־uv בוחר אותה בעצמו.

> **טעות נפוצה:** להריץ pip „רק הפעם”. ב־Ubuntu החדשות זה ממילא נכשל בשגיאה
> `externally-managed-environment`, ובפרויקט זה עוקף את קובץ הנעילה. מוסיפים חבילה עם `uv add`, ומריצים
> עם `uv run`.

**הפעלה מחדש ושדרוג.** אין שירות. כדי לעבור לגרסה מוצמדת אחרת, חדשה או ישנה, מריצים שוב את פקודת ההתקנה
עם מספר הגרסה הרצוי.

**הסרה.** לפי התיעוד של uv: מנקים מטמון, מוחקים את גרסאות ה־Python ואת הכלים שהותקנו דרכו, ואז את שני
קובצי ההרצה:

<div dir="ltr" align="left">

```bash
uv cache clean
rm -r "$(uv python dir)" "$(uv tool dir)"
rm ~/.local/bin/uv ~/.local/bin/uvx ~/.local/bin/python3.14
```

</div>

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| `uv: command not found` | לשונית שנפתחה לפני ההתקנה | לשונית חדשה, או `source ~/.bashrc` |
| `No interpreter found for Python 3.14` | Python 3.14 לא הותקנה דרך uv | `uv python install 3.14` |
| `externally-managed-environment` | הורצה pip על ה־Python של המערכת | `uv add` ו־`uv run`; לא pip |
| `ModuleNotFoundError` כשמריצים `python3 ...` | ה־Python של המערכת, בלי הסביבה של הפרויקט | `uv run python ...` מתוך תיקיית הפרויקט |

בזה נגמר יום 0. פתחי לשונית חדשה והריצי את בדיקות הגרסה של כל כלי שהתקנת; ביום 3 הסקריפט `doctor.sh` יעשה
את זה בשבילך.

## 2.9 מאגר הקורס ובדיקת עשן

זה הסעיף של יום 3. שתי תיקיות ילוו אותך עד סוף הספר:

| תיקייה | מה בה | משתנה |
| --- | --- | --- |
| `~/senior-software-engineer-course` | הספר עצמו: פרקים, תרגילים, פתרונות, כלים. קוראים ממנו ומעתיקים ממנו. | `COURSE` |
| `~/senior-course` | העבודה שלך: מאגר Git אחד, תיקייה לכל פרק (`ch01`, `ch02` וכן הלאה) | |

**הספר בתוך WSL.** הדרך תלויה במקום שבו הספר נמצא אצלך:

| איפה הספר | איך מביאים אותו |
| --- | --- |
| במאגר שלך ב־GitHub | `git clone` בכתובת ה־SSH שלו, לתוך `~/senior-software-engineer-course` |
| כקובץ zip שהורדת ב־Windows | `unzip` מתוך התיקייה של Windows, לתוך הבית שלך ב־Linux |

שתי האפשרויות, בפקודות; הריצי רק את המתאימה, אחרי שתחליפי את השם בסוגריים המשולשים:

<div dir="ltr" align="left">

```bash
git clone git@github.com:<your-user>/senior-software-engineer-course.git ~/senior-software-engineer-course
unzip /mnt/c/Users/<WindowsUser>/Downloads/senior-software-engineer-course.zip -d ~
```

</div>

זה השימוש הלגיטימי היחיד ב־`/mnt/c` בספר: מעתיקים פעם אחת מ־Windows לתוך Linux, ומשם עובדים. עכשיו המשתנה
`COURSE`, שכל README של תרגיל משתמש בו:

<div dir="ltr" align="left">

```bash
echo 'export COURSE="$HOME/senior-software-engineer-course"' >> ~/.bashrc
source ~/.bashrc
ls "$COURSE"
```

</div>

<div dir="ltr" align="left">

```text
00-stage1  README.md  book  exercises  obsidian  site  tools
```

</div>

**המאגר שלך.** ניצור את `~/senior-course`, נעתיק אליו קובץ החרגות מוכן (מה ש־Git לא יעקוב אחריו: סביבות
וירטואליות, קבצי בנייה, סודות), ואת קוד ההתחלה של הפרק:

<div dir="ltr" align="left">

```bash
mkdir -p ~/senior-course
cd ~/senior-course
git init
cp "$COURSE/exercises/ch02-workbench/data/gitignore" .gitignore
mkdir -p ch02
cp -r "$COURSE/exercises/ch02-workbench/starter/." ch02/
ls ch02
```

</div>

<div dir="ltr" align="left">

```text
doctor.sh  hello  smoke  test_doctor.sh
```

</div>

### 2.9.1 הרופא: `doctor.sh`

הסקריפט בודק כל כלי מיום 0: שהוא קיים, ושהגרסה שלו לפחות המינימום. הוא בודק גם את המשתנה `COURSE`, את
המאגר שלך, ושאף אחד מהם לא יושב על כונן של Windows.

<div dir="ltr" align="left">

```bash
bash ch02/doctor.sh
echo "exit code: $?"
```

</div>

<div dir="ltr" align="left">

```text
doctor: checking the course workbench on Linux x86_64 (WSL2)
Day 0 tools:
  ok    git                2.53.0       (need >= 2.28)
  ok    ssh-key            found        (a public key in ~/.ssh)
  ok    docker             29.8.1       (need >= 29.0)
  ok    docker-engine      29.8.1       (need >= 29.0)
  ok    compose            ...          (need >= 2.20)
  ok    code               ...          (any version)
  ok    go                 1.27.1       (need >= 1.27.1)
  ok    uv                 0.12.23      (need >= 0.12.23)
  ok    python             3.14.8       (need >= 3.14.0)
Go tools (go install, Day 0):
  ok    staticcheck        2026.2.1     (need >= 2026.2.1)
  ok    goose              3.28.0       (need >= 3.28.0)
  ok    buf                1.73.0       (need >= 1.73.0)
  ok    protoc-gen-go      1.36.12      (need >= 1.36.12)
  ok    protoc-gen-go-grpc 1.6.2        (need >= 1.6.2)
  ok    govulncheck        1.8.0        (need >= 1.8.0)
  ok    actionlint         1.7.12       (need >= 1.7.12)
  ok    golangci-lint      2.14.0       (need >= 2.14.0)
Course folders:
  ok    COURSE             found        (/home/noga/senior-software-engineer-course)
  ok    senior-course      git-repo     (/home/noga/senior-course)
doctor: 19 ok, 0 failed
doctor: the workbench is ready
exit code: 0
```

</div>

זו צורת הפלט כשהכול תקין; המספרים אצלך יהיו מה שהתקנת. שורת `FAIL` אומרת מה נמצא, מה נדרש, ואיזה סעיף
מתקין את הכלי. קוד היציאה של הסקריפט הוא **מספר הבדיקות שנכשלו**. זו החלטה מכוונת: 0 אומר „מוכן”, וכל
סקריפט אחר, או צינור CI, יכול להחליט לפי המספר בלי לקרוא את הטקסט. פרויקט הגמר משתמש בסקריפט הזה בדיוק.

שתי פונקציות עושות את רוב העבודה. הראשונה, מתוך `doctor.sh`, משווה גרסאות כמספרים, שדה אחר שדה:

<div dir="ltr" align="left">

```bash
# version_ge A B: succeed when version A >= version B, comparing three numeric fields.
version_ge() {
  local IFS=.
  local -a a b
  local i x y
  read -r -a a <<< "$1"
  read -r -a b <<< "$2"
  for i in 0 1 2; do
    x=${a[i]:-0}
    y=${b[i]:-0}
    if ((10#$x > 10#$y)); then return 0; fi
    if ((10#$x < 10#$y)); then return 1; fi
  done
  return 0
}
```

</div>

למה לא להשוות מחרוזות? כי בהשוואת מחרוזות, הגרסה `2.10.0` קטנה מ־`2.3.0`: התו הראשון שונה הוא 1 מול 3.
הקידומת `10#` מכריחה קריאה בבסיס עשר, כדי ששדה כמו `08` לא ייקרא כמספר אוקטלי.

הפונקציה השנייה, `check_tool`, מקבלת שם, גרסה מינימלית, הפניה לסעיף, ואת פקודת הגרסה של הכלי. היא בודקת שהפקודה קיימת, מריצה אותה, שולפת את המספר הראשון בפלט, ומשווה.

הסקריפט נכתב כך שירוץ גם ב־bash 3.2, הגרסה הישנה שמגיעה עם macOS, ולכן אין בו מערכים אסוציאטיביים
ותחביר חדש אחר. לצידו נמצא `test_doctor.sh`, שמריץ אותו מול כלים מזויפים (סקריפטים קטנים שמדפיסים גרסה)
ובודק את קודי היציאה. ביום 3 שתי בדיקות שלו נכשלות, וזה צפוי: הן בודקות את הכלים של חלק ה, שתוסיפי ביום 57.

### 2.9.2 בדיקת עשן ל־Go

בדיקת עשן (smoke test) היא בדיקה קטנה שמוכיחה שהשרשרת כולה עובדת: מהדר, בדיקות והרצה. המודול
`ch02/hello` מכיל פונקציה אחת. הבדיקה שלו היא בדיקה טבלאית (table-driven test): טבלה של קלטים ופלטים
צפויים, ולולאה שמריצה כל שורה כתת־בדיקה עם שם משלה. זה הקובץ `main_test.go`:

<div dir="ltr" align="left">

```go
package main

import "testing"

func TestGreeting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain name", "Noga", "Hello, Noga!"},
		{"two words", "Noga M", "Hello, Noga M!"},
		{"surrounding spaces", "  Noga ", "Hello, Noga!"},
		{"empty", "", "Hello, world!"},
		{"only spaces", "   ", "Hello, world!"},
		{"hebrew", "נגה", "Hello, נגה!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := greeting(tt.in); got != tt.want {
				t.Errorf("greeting(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
```

</div>

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch02/hello
go test ./...
```

</div>

<div dir="ltr" align="left">

```text
--- FAIL: TestGreeting (0.00s)
    --- FAIL: TestGreeting/surrounding_spaces (0.00s)
        main_test.go:21: greeting("  Noga ") = "Hello,   Noga !", want "Hello, Noga!"
    --- FAIL: TestGreeting/empty (0.00s)
        main_test.go:21: greeting("") = "Hello, !", want "Hello, world!"
    --- FAIL: TestGreeting/only_spaces (0.00s)
        main_test.go:21: greeting("   ") = "Hello,    !", want "Hello, world!"
FAIL
```

</div>

שלוש שורות אדומות, וזה בכוונה: הן מראות איך נראה כישלון של בדיקה טבלאית, עם השם של השורה שנכשלה והערכים.
פתחי את `main.go` ומלאי את ה־`TODO`. הפקודה `code .` פותחת את התיקייה הנוכחית בעורך. זה הפתרון, מתוך
`solution/hello/main.go`:

<div dir="ltr" align="left">

```go
// greeting returns "Hello, <name>!", or "Hello, world!" when name is empty or only spaces.
func greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "world"
	}
	return "Hello, " + name + "!"
}
```

</div>

<div dir="ltr" align="left">

```bash
go test ./...
go run . Noga
```

</div>

<div dir="ltr" align="left">

```text
ok  	example.com/ch02-workbench/hello	0.002s
Hello, Noga!
```

</div>

> **איך Rust עושה את זה:** ב־Rust היית כותבת פונקציית `#[test]` עם לולאה על מערך של זוגות, או משתמשת בספריית
> בדיקות פרמטריות. ההבדל ב־Go הוא `t.Run`: כל שורה בטבלה הופכת לתת־בדיקה עם שם, כך שהכישלון מדווח בדיוק איזו
> שורה נפלה, ואפשר להריץ שורה אחת לבד עם הדגל `-run`. פרק 3 משתמש בצורה הזאת בכל תרגיל.

### 2.9.3 בדיקת עשן ל־Python

התיקייה `ch02/smoke` היא פרויקט uv קטן. בדיקות העשן שלו, מתוך `tests/test_smoke.py`:

<div dir="ltr" align="left">

```python
"""Smoke tests: the right Python, inside the project environment, importing the project."""

import sys

from smoke import describe


def test_python_is_3_14() -> None:
    assert sys.version_info[:2] == (3, 14)


def test_runs_inside_the_project_environment() -> None:
    # `uv run` puts the project's .venv first; a stray system Python fails here.
    assert sys.prefix != sys.base_prefix


def test_describe_names_version_and_system() -> None:
    text = describe()
    assert text.startswith("Python 3.14.")
    assert " on " in text
```

</div>

הבדיקה השנייה היא החשובה: היא נכשלת אם הבדיקות רצות ב־Python של המערכת ולא בסביבה הווירטואלית של
הפרויקט. ההבדל בין שני הערכים:

| ערך | בתוך סביבה וירטואלית |
| --- | --- |
| `sys.prefix` | התיקייה `.venv` של הפרויקט |
| `sys.base_prefix` | ה־Python שממנו הסביבה נוצרה |

<div dir="ltr" align="left">

```bash
cd ~/senior-course/ch02/smoke
uv sync
uv run pytest -q
```

</div>

הפקודה `uv sync` יוצרת את `.venv` ומתקינה בדיוק את הגרסאות שבקובץ הנעילה. הבדיקה השלישית נכשלת עד שתמלאי
את ה־`TODO` בקובץ `src/smoke/__init__.py`. זה הפתרון, מתוך `solution/smoke/src/smoke/__init__.py`:

<div dir="ltr" align="left">

```python
"""Smoke test for the Python toolchain (chapter 2)."""

import platform
import sys


def describe() -> str:
    """Return the running interpreter in one line, for example "Python 3.14.8 on Linux"."""
    v = sys.version_info
    return f"Python {v.major}.{v.minor}.{v.micro} on {platform.system()}"
```

</div>

<div dir="ltr" align="left">

```text
...                                                                      [100%]
3 passed in 0.01s
```

</div>

### 2.9.4 הקונטיינר הראשון וההתחייבות הראשונה

<div dir="ltr" align="left">

```bash
docker run --rm hello-world
```

</div>

<div dir="ltr" align="left">

```text
Hello from Docker!
This message shows that your installation appears to be working correctly.
...
```

</div>

הפקודה מורידה אימג' (image) קטן מהמאגר הציבורי של Docker, יוצרת ממנו קונטיינר ומריצה אותו. הדגל `--rm`
מוחק את הקונטיינר כשהוא מסיים. ההודעה המלאה מפרטת את ארבעת הצעדים האלה; קראי אותה, היא תשובה טובה לשאלה „מה
Docker עושה כשאני מריצה קונטיינר”.

ועכשיו ההתחייבות. היא כוללת גם את התיקייה `ch01` מתרגיל פרק 1, אם עוד לא התחייבת עליה. הפקודה הראשונה למטה מציגה
מה ייכנס. ודאי שאין ברשימה `.venv` ואין קבצי בנייה; אם יש, קובץ ההחרגות לא הועתק:

<div dir="ltr" align="left">

```bash
cd ~/senior-course
git status --short
git add .
git commit -m "ch01-ch02: postings tally, workbench, doctor and smoke tests"
git log --oneline
```

</div>

<div dir="ltr" align="left">

```text
2ac1ae6 ch01-ch02: postings tally, workbench, doctor and smoke tests
```

</div>

המזהה בתחילת השורה יהיה אחר אצלך. מכאן והלאה, כל פרק מסתיים בהתחייבות כזאת.

## 2.10 התקנות של חלק ה: kind, kubectl, Helm, Terraform, k6

זה הסעיף של יום 57. לא יוצרים היום אשכול ולא מריצים בדיקת עומס; רק מתקינים, בודקים ומרחיבים את `doctor.sh`.
פרק 20 מתחיל להשתמש בכלים ביום 60.

| כלי | מה הוא | פרק | גרסה, נכון לאוקטובר 2026 (נספח ג) |
| --- | --- | --- | --- |
| kind | אשכול Kubernetes שלם בתוך קונטיינרים של Docker | 20 | v0.33.0 |
| kubectl | שורת הפקודה של Kubernetes | 20 | v1.37.1 |
| Helm | מנהל חבילות ל־Kubernetes (charts) | 21 | v4.3.0 |
| Terraform | תשתית כקוד (infrastructure as code) | 22 | 1.16.5 |
| k6 | בדיקות עומס (load tests) בסקריפטים | 26 | v2.3.0 |

כל החמישה הם קובץ הרצה יחיד, והם נכנסים ל־`~/.local/bin` בלי `sudo`. לכל אחד אותו דפוס: מורידים את הקובץ ואת
קובץ הטביעות מאותו פרסום, בודקים, ורק אז מתקינים. הפקודה `install -m 0755` מעתיקה את הקובץ ומסמנת אותו
כקובץ הרצה. מתחילים בתיקייה זמנית:

<div dir="ltr" align="left">

```bash
ARCH=$(dpkg --print-architecture)
mkdir -p /tmp/part-e && cd /tmp/part-e
```

</div>

> **שימי לב:** טביעה שמגיעה מאותו שרת כמו הקובץ מוכיחה שההורדה שלמה ולא נפגמה בדרך. היא לא מוכיחה מי
> פרסם את הקובץ; לזה צריך חתימה קריפטוגרפית, ופרק 28 חוזר לנושא שרשרת האספקה.

**kind.**

<div dir="ltr" align="left">

```bash
curl -fsSLO "https://github.com/kubernetes-sigs/kind/releases/download/v0.33.0/kind-linux-${ARCH}"
curl -fsSLO "https://github.com/kubernetes-sigs/kind/releases/download/v0.33.0/kind-linux-${ARCH}.sha256sum"
sha256sum -c "kind-linux-${ARCH}.sha256sum"
install -m 0755 "kind-linux-${ARCH}" ~/.local/bin/kind
kind version
```

</div>

<div dir="ltr" align="left">

```text
kind-linux-amd64: OK
kind v0.33.0 go1.26.7 linux/amd64
```

</div>

הכלי kind מריץ כל node (שרת אחד באשכול) כקונטיינר, שנבנה מאימג' מיוחד. שם האימג' הוא `kindest/node`. הוראות הפרסום של kind v0.33.0
אומרות שחובה להצמיד את האימג' לפי הטביעה שלו (digest) ולא רק לפי התגית, כי תגית יכולה לזוז והטביעה לא.
נוריד אותו כבר היום, כדי שיום 60 לא יתחיל בהורדה ארוכה:

<div dir="ltr" align="left">

```bash
docker pull kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5
docker image ls --digests kindest/node
```

</div>

**kubectl.** כאן קובץ הטביעה מכיל רק את הטביעה, בלי שם קובץ, ולכן מרכיבים את השורה בעצמנו:

<div dir="ltr" align="left">

```bash
curl -fsSLO "https://dl.k8s.io/release/v1.37.1/bin/linux/${ARCH}/kubectl"
curl -fsSLO "https://dl.k8s.io/release/v1.37.1/bin/linux/${ARCH}/kubectl.sha256"
echo "$(cat kubectl.sha256)  kubectl" | sha256sum -c -
install -m 0755 kubectl ~/.local/bin/kubectl
kubectl version --client
```

</div>

<div dir="ltr" align="left">

```text
kubectl: OK
Client Version: v1.37.1
Kustomize Version: v5.8.1
```

</div>

הדגל `--client` מבקש רק את גרסת הכלי. בלעדיו kubectl ינסה להגיע לאשכול, שעוד לא קיים, וידפיס שגיאה מבלבלת.

**Helm.**

<div dir="ltr" align="left">

```bash
curl -fsSLO "https://get.helm.sh/helm-v4.3.0-linux-${ARCH}.tar.gz"
curl -fsSLO "https://get.helm.sh/helm-v4.3.0-linux-${ARCH}.tar.gz.sha256sum"
sha256sum -c "helm-v4.3.0-linux-${ARCH}.tar.gz.sha256sum"
tar -xzf "helm-v4.3.0-linux-${ARCH}.tar.gz"
install -m 0755 "linux-${ARCH}/helm" ~/.local/bin/helm
helm version --short
```

</div>

<div dir="ltr" align="left">

```text
helm-v4.3.0-linux-amd64.tar.gz: OK
v4.3.0+gbec5b06
```

</div>

**Terraform.** הקובץ של HashiCorp מכיל טביעות לכל מערכות ההפעלה, והדגל `--ignore-missing` אומר לבדוק רק את
הקבצים שבאמת הורדת:

<div dir="ltr" align="left">

```bash
curl -fsSLO "https://releases.hashicorp.com/terraform/1.16.5/terraform_1.16.5_linux_${ARCH}.zip"
curl -fsSLO "https://releases.hashicorp.com/terraform/1.16.5/terraform_1.16.5_SHA256SUMS"
sha256sum -c --ignore-missing terraform_1.16.5_SHA256SUMS
unzip -o -q "terraform_1.16.5_linux_${ARCH}.zip" terraform
install -m 0755 terraform ~/.local/bin/terraform
terraform version
```

</div>

<div dir="ltr" align="left">

```text
terraform_1.16.5_linux_amd64.zip: OK
Terraform v1.16.5
on linux_amd64
```

</div>

**k6.**

<div dir="ltr" align="left">

```bash
curl -fsSLO "https://github.com/grafana/k6/releases/download/v2.3.0/k6-v2.3.0-linux-${ARCH}.tar.gz"
curl -fsSLO "https://github.com/grafana/k6/releases/download/v2.3.0/k6-v2.3.0-checksums.txt"
sha256sum -c --ignore-missing k6-v2.3.0-checksums.txt
tar -xzf "k6-v2.3.0-linux-${ARCH}.tar.gz"
install -m 0755 "k6-v2.3.0-linux-${ARCH}/k6" ~/.local/bin/k6
k6 version
```

</div>

<div dir="ltr" align="left">

```text
k6-v2.3.0-linux-amd64.tar.gz: OK
k6 v2.3.0 (commit/e088784614, go1.26.8, linux/amd64)
```

</div>

הפלטים בסעיף הזה הועתקו מהרצה בקונטיינר Ubuntu 26.04 על מעבד `arm64`; אצלך תופיע הארכיטקטורה שלך, ושאר
הטקסט זהה.

**הרחבת `doctor.sh`.** ב־`ch02/doctor.sh` יש בלוק של חלק ה עם `TODO`. הוסיפי בו שורת `check_tool` לכל כלי, לפי
הדפוס של שורות יום 0. זה הבלוק המלא, מתוך `solution/doctor.sh`:

<div dir="ltr" align="left">

```bash
if [ "$PART_E" -eq 1 ]; then
  echo "Part E tools:"
  check_tool "kind" "0.33.0" "chapter 2, section 2.10" kind version
  check_tool "kubectl" "1.37.1" "chapter 2, section 2.10" kubectl version --client
  check_tool "helm" "4.3.0" "chapter 2, section 2.10" helm version --short
  check_tool "terraform" "1.16.5" "chapter 2, section 2.10" terraform version
  check_tool "k6" "2.3.0" "chapter 2, section 2.10" k6 version
fi
```

</div>

<div dir="ltr" align="left">

```bash
cd ~/senior-course
bash ch02/doctor.sh --part-e
bash ch02/test_doctor.sh
```

</div>

הפעם הסקריפט מסתיים ב־`0 failed`, והבדיקות שלו מסתיימות ב־11 הצלחות. Terraform נוהג לבדוק ברשת אם
יצאה גרסה חדשה; הסקריפט מכבה את זה עם המשתנה `CHECKPOINT_DISABLE`.

**שדרוג והסרה.** שדרוג הוא אותן פקודות עם גרסה חדשה. הסרה היא מחיקת הקבצים, ואם תרצי גם את ההגדרות
והמטמונים שהכלים יוצרים בשימוש:

<div dir="ltr" align="left">

```bash
rm ~/.local/bin/kind ~/.local/bin/kubectl ~/.local/bin/helm ~/.local/bin/terraform ~/.local/bin/k6
rm -rf ~/.kube ~/.cache/helm ~/.config/helm ~/.local/share/helm ~/.terraform.d
rm -rf /tmp/part-e
```

</div>

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| `cannot execute binary file: Exec format error` | הורד קובץ של ארכיטקטורה אחרת | לבדוק את המשתנה `ARCH` ולהוריד שוב |
| `FAILED` בבדיקת הטביעה | הורדה חלקית, או גרסה שונה בין שני הקבצים | למחוק את שניהם ולהוריד שוב |
| kubectl מדפיס `connection refused` | נשכח הדגל `--client` | אין אשכול עד יום 60; זה תקין |
| `doctor.sh --part-e` מראה גרסה ישנה | עותק ישן של הכלי מוקדם ב־`PATH` | `command -v <כלי>`, ולמחוק את העותק הישן |

## 2.11 פתרון תקלות

רוב התקלות בהתקנה הן אחת משלוש: הפקודה לא נמצאת, נמצאת גרסה אחרת ממה שהתקנת, או שאת בתיקייה הלא נכונה.
סדר הבדיקה שפותר את רובן, לפני שמחפשים ברשת:

1. **קראי את השורה האחרונה של השגיאה, עד הסוף.** ברוב המקרים היא אומרת בדיוק מה חסר.
2. **איפה את?** הפקודה `pwd`. אם התשובה מתחילה ב־`/mnt/c`, עברי ל־`~`.
3. **מי ירוץ?** הפקודה `command -v <כלי>` מראה את הנתיב. נתיב תחת `/usr/bin` כשציפית ל־`~/.local/bin`
   אומר שעותק של המערכת מנצח.
4. **ה־`PATH` עדכני?** פתחי לשונית חדשה, או הריצי `source ~/.bashrc`, ונסי שוב.
5. **הרופא.** `bash ~/senior-course/ch02/doctor.sh` מסכם את המצב בעמוד אחד, עם הסעיף שמתקן כל שורה.
6. **נספח ב** מרכז את התקלות הידועות לפי הודעת השגיאה, כולל תקלות שאופייניות ל־Mac.

| מה רואים | הסיבה הסבירה | מה עושים |
| --- | --- | --- |
| הכול איטי מאוד, ושעונים של קבצים לא מתעדכנים | העבודה נעשית תחת `/mnt/c` | להעביר את התיקייה אל `~` |
| `doctor.sh` עובר אבל VS Code מראה שגיאות | VS Code נפתח מצד Windows | `code .` מתוך Ubuntu |
| הורדות נכשלות בשגיאת תעודה | רשת ארגונית עם proxy | לנסות מרשת ביתית; לא לבטל אימות תעודות |
| המחשב איטי כש־Docker רץ | WSL ו־Docker תופסים זיכרון רב | סעיף 2.12, הגבלת זיכרון |
| שינוי ב־`~/.bashrc` שבר את הטרמינל | שגיאת תחביר בקובץ | `bash --norc`, ואז לתקן את הקובץ ב־`nano ~/.bashrc` |

אם אחרי 30 דקות תקלה עדיין עומדת, רשמי ביומן התקיעות (פרק 0, סעיף 0.9) מה ניסית ומה ראית, והמשיכי למשהו
אחר. התקנה שנתקעה היא בדיוק המקום שבו יום הפסקה עוזר.

## 2.12 עצירה, הפעלה מחדש וניקוי

**ביום רגיל.** אין צורך לעצור כלום בין ימי לימוד, אבל שלושה הרגלים חוסכים זיכרון ודיסק:

| מתי | מה עושים |
| --- | --- |
| בסוף תרגיל עם compose | `docker compose -p chNN down` בתיקיית התרגיל, ו־`down -v` אם רוצים גם למחוק את הנתונים |
| בסוף יום שבו לא צריך קונטיינרים | לצאת מ־Docker Desktop |
| כשהמחשב איטי או WSL מתנהג מוזר | `wsl --shutdown` ב־PowerShell, ולפתוח טרמינל מחדש |

**מקום בדיסק של Docker.** אימג'ים, קונטיינרים שעצרו, ונפחים (volumes) מצטברים. קודם מסתכלים, אחר כך מוחקים:

<div dir="ltr" align="left">

```bash
docker system df
docker container prune
docker image prune
docker system prune
docker volume ls
```

</div>

| פקודה | מה היא מוחקת |
| --- | --- |
| `docker system df` | כלום; מציגה כמה מקום תופס כל סוג |
| `docker container prune` | קונטיינרים שעצרו |
| `docker image prune` | אימג'ים שאין להם תגית (שכבות ישנות של בניות) |
| `docker system prune` | את שני הקודמים, ועוד רשתות שאינן בשימוש ומטמון בנייה |
| `docker volume ls` | כלום; מציגה את הנפחים, שבהם חיים נתוני PostgreSQL ו־Redis של התרגילים |

כל הפקודות שמוחקות שואלות לפני כן. נפחים לא נמחקים באף אחת מהן, בכוונה: שם נמצאים הנתונים שלך. מחיקה של
נפח היא פעולה מפורשת, בדגל `-v` של `down` בתרגיל, או בפקודה שמתחת:

<div dir="ltr" align="left">

```bash
docker volume rm <name>
```

</div>

**הדיסק הווירטואלי לא קטן מעצמו.** כל הפצה של WSL שמורה בקובץ `ext4.vhdx` שגדל לפי הצורך. לפי התיעוד של
Microsoft, קובץ כזה גדל כשמוסיפים קבצים, אבל לא קטן כשמוחקים אותם. אחרי ניקוי גדול בתוך Ubuntu, הנה הדרך
להחזיר את המקום ל־Windows. הצעדים לא הורצו בהרצה שבה נכתב הספר (אין שם Windows); הם מבוססים על שני
עמודי התיעוד שבסוף הסעיף.

1. **עצירה.** צאי מ־Docker Desktop, ואז ב־PowerShell עצרי את WSL. הקובץ חייב להיות מנותק בזמן הכיווץ.
2. **הנתיב לקובץ.** ב־PowerShell רגיל, הפקודה השנייה למטה מדפיסה את הנתיב המלא של הדיסק של Ubuntu. העתיקי אותו.
3. **כיווץ.** פתחי PowerShell עם הרשאות מנהלת והריצי `diskpart`. בשורה `DISKPART>` הקלידי את שלוש השורות של
   הבלוק השני, עם הנתיב שהעתקת. הכיווץ עשוי לקחת כמה דקות.

<div dir="ltr" align="left">

```powershell
wsl --shutdown
(Get-ChildItem -Path HKCU:\Software\Microsoft\Windows\CurrentVersion\Lxss | Where-Object { $_.GetValue("DistributionName") -eq 'Ubuntu' }).GetValue("BasePath") + "\ext4.vhdx"
```

</div>

<div dir="ltr" align="left">

```text
select vdisk file="<path-to-ext4.vhdx>"
compact vdisk
exit
```

</div>

שימי לב לשני דברים. ראשית, Microsoft ממליצה לא לגעת בקבצים של WSL בכלים של Windows, חוץ מהפעולות המתועדות
האלה; לא מזיזים את הקובץ ולא פותחים אותו. שנית, האימג'ים של Docker לא יושבים בדיסק של Ubuntu אלא בדיסק
נפרד של Docker Desktop. הצעדים כאן לא מטפלים בו, ולכן ייתכן שהמקום ש־`docker system prune` פינה לא יחזור
ל־Windows. לדיסק הזה יש הגדרות משלו ב־Docker Desktop, ותיעוד משלו באתר של Docker.

<div dir="ltr" align="left">

```text
https://learn.microsoft.com/en-us/windows/wsl/disk-space
https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/compact-vdisk
```

</div>

**הגבלת הזיכרון של WSL.** לפי התיעוד של Microsoft, ברירת המחדל היא חצי מהזיכרון של המחשב. אם המחשב נעשה
איטי כשאשכול kind וכמה קונטיינרים רצים, אפשר להגביל:

- **הקובץ:** קובץ טקסט בשם `.wslconfig`, ב־Windows, בתיקייה שהמשתנה `%UserProfile%` מצביע עליה.
- **התוכן:** הבלוק שמתחת. המספרים הם נקודת פתיחה; התאימי למחשב שלך.
- **הפעלה:** אחרי השמירה, עצירה של WSL ב־PowerShell (כמו בצעד הראשון למעלה), כדי שההגדרה תיקלט.

<div dir="ltr" align="left">

```text
[wsl2]
memory=8GB
processors=4
```

</div>

**ניקוי מלא.** אם תרצי להתחיל מאפס: כל סעיף בפרק כולל את פקודות ההסרה של הכלי שלו, בסדר הפוך להתקנה. המאגר
שלך ב־`~/senior-course` שווה יותר מכל כלי. לפני `--unregister` (סעיף 2.2), ודאי שהוא נשמר ב־GitHub, או
בקובץ הגיבוי של `wsl --export`.

## 2.13 אם את עובדת על Mac

על Mac אין WSL: macOS עצמה היא מערכת Unix, והטרמינל שלה מריץ כמעט את כל הפקודות בספר כמו שהן. ההבדלים:

| נושא | ב־WSL2 (בספר) | ב־Mac |
| --- | --- | --- |
| מערכת | Ubuntu בתוך WSL2 | macOS עצמה; אין שכבה נוספת |
| טרמינל ומעטפת | Windows Terminal ו־bash | Terminal עם zsh כברירת מחדל |
| קובץ ההגדרות של המעטפת | `~/.bashrc` | `~/.zshrc` (את שורות ה־`PATH` וה־`COURSE` כותבים שם) |
| חבילות מערכת | `sudo apt install ...` | Homebrew: `brew install ...` |
| Docker | Docker Desktop עם WSL2 | Docker Desktop for Mac, הגרסה המתאימה למעבד |
| Git | `apt` | מגיע עם כלי הפיתוח של Xcode (`xcode-select --install`), או Homebrew |
| VS Code | הרחבת WSL | אין צורך בהרחבה; הפקודה `code` מותקנת מתוך לוח הפקודות |
| Go | `go1.27.1.linux-amd64.tar.gz` | `go1.27.1.darwin-arm64.tar.gz` (או `amd64` במחשב Intel) |
| ארכיטקטורה | `ARCH=$(dpkg --print-architecture)` | `ARCH=arm64` (מעבד של Apple), או `ARCH=amd64` (מעבד Intel) |
| kind ו־kubectl | `kind-linux-...`, `bin/linux/...` | `kind-darwin-...`, `bin/darwin/...` |
| Helm ו־Terraform | `linux-${ARCH}`, `linux_${ARCH}` | `darwin-${ARCH}`, `darwin_${ARCH}` |
| k6 | `k6-v2.3.0-linux-${ARCH}.tar.gz` | `k6-v2.3.0-macos-${ARCH}.zip`, נפתח עם `unzip` |
| golangci-lint | `golangci-lint-2.14.0-linux-${ARCH}` | `golangci-lint-2.14.0-darwin-${ARCH}` |
| מפתח SSH בזיכרון | `keychain` | `ssh-add --apple-use-keychain ~/.ssh/id_ed25519` |

> **ב־Mac:** הפקודה `sed` שונה במקום אחד. בגרסה של macOS הדגל `-i` דורש ארגומנט, גם אם הוא ריק. זה ההבדל
> היחיד בפקודות הספר שגורם לשגיאה מבלבלת (`invalid command code`) במקום לעבוד:

<div dir="ltr" align="left">

```bash
sed -i 's/a/b/' file      # Ubuntu
sed -i '' 's/a/b/' file   # macOS
```

</div>

פקודות ההורדה ובדיקת הטביעות של סעיפים 2.7, 2.7.1 ו־2.10 עובדות ב־Mac כמו שהן, אחרי החלפת השמות לפי הטבלה. הן נבדקו
בפועל על Mac עם מעבד של Apple. שתי הפקודות `sha256sum` ו־`unzip` קיימות במערכת. גם הפקודה `shasum -a 256 -c`
עובדת. לבדיקת הטביעה של Go צריך גם את `jq`:

<div dir="ltr" align="left">

```bash
brew install jq
```

</div>

שתי הערות על המעטפת. ראשית, הסקריפטים של הספר נכתבו ל־bash, ומריצים אותם במפורש דרכה; לכן המעטפת השונה
של Mac לא מפריעה. הגרסה של bash שמגיעה עם macOS היא 3.2 הישנה, והרופא נבדק גם בה. שנית, קבצים שהורדו בדפדפן
מסומנים ב־macOS כ„הגיעו מהאינטרנט”, והמערכת עלולה לחסום את הרצתם; קבצים שהורדו ב־`curl` לא מסומנים כך.
אם הורדת בדפדפן ונחסמת, הפקודה שמתחת מסירה את הסימון:

<div dir="ltr" align="left">

```bash
xattr -d com.apple.quarantine <file>
```

</div>

את Docker Desktop מורידים מעמוד ההתקנה הרשמי, בגרסה שמתאימה למעבד; הדרישות המעודכנות מופיעות שם.
כל השאר בפרק, מסעיף 2.9 והלאה, זהה: אותן תיקיות, אותו `doctor.sh`, אותן בדיקות.

## סיכום

- מערכת ההפעלה: Ubuntu בתוך WSL2.
- קונטיינרים: Docker Desktop, על אותו מנוע וירטואלי.
- ניהול גרסאות: Git, עם מפתח SSH להזדהות. העורך: VS Code, שעובד על הקבצים שבתוך המערכת.
- שפות: Go ו־uv בתיקיית הבית; ביום 57 נוספים חמשת הכלים של חלק ה.
- עובדים רק בתוך מערכת הקבצים של Linux, תחת `~`; הכוננים של Windows תחת `/mnt/c` הם לייבוא חד־פעמי.
- הכלים של הקורס נכנסים לתיקיית הבית, בלי `sudo`. תיקייה מוקדמת ב־`PATH` מנצחת, ואחרי שינוי פותחים לשונית חדשה.
- כל הורדה נבדקת מול הטביעה שפורסמה לצידה, לפני ההתקנה. כל אימג' שחייב להיות זהה מוצמד לפי טביעה.
- `doctor.sh` מסכם את המצב, וקוד היציאה שלו הוא מספר הכשלים; הוא נשען על השוואת גרסאות כמספרים.
- בדיקות העשן מוכיחות שהשרשרת כולה עובדת: בדיקה טבלאית ב־Go, pytest בסביבה הווירטואלית, וקונטיינר ראשון.
- לכל כלי יש דרך לעצור, לנקות ולהסיר; נפחים של Docker נמחקים רק במפורש.

## מונחים לזכור

| מונח בעברית | באנגלית | במשפט אחד |
| --- | --- | --- |
| תת־מערכת Linux ל־Windows | WSL2 | ליבת Linux אמיתית במכונה וירטואלית קלה, עם הפצה כמו Ubuntu מעליה |
| מעטפת | Shell | התוכנית שקוראת פקודות ומריצה אותן; ב־Ubuntu זו bash |
| משתנה הסביבה PATH | PATH | רשימת התיקיות שבהן המעטפת מחפשת פקודה, לפי הסדר |
| טביעת קובץ | Checksum (SHA-256) | מחרוזת שמחושבת מתוכן הקובץ; זהות שלה מוכיחה שההורדה שלמה |
| טביעת אימג' | Image digest | מזהה של אימג' לפי התוכן שלו; בניגוד לתגית, הוא לא זז |
| מפתח SSH | SSH key pair | מפתח פרטי שנשאר במחשב, ומפתח ציבורי שנותנים לשרת |
| בדיקת עשן | Smoke test | בדיקה קטנה שמוכיחה שהשרשרת כולה עובדת |
| בדיקה טבלאית | Table-driven test | טבלה של קלטים ופלטים צפויים, ולולאה שמריצה כל שורה כתת־בדיקה |
| סביבה וירטואלית | Virtual environment | התיקייה `.venv` עם Python והחבילות של פרויקט אחד |
| קוד יציאה | Exit code | המספר שתוכנית מחזירה בסיומה; 0 פירושו הצלחה |

## טעויות נפוצות

- **לעבוד תחת `/mnt/c` „כי שם הקבצים”.** הכול איטי, והרשאות וסופי שורות משתבשים. מעתיקים פעם אחת לתוך `~`.
- **להתקין Docker גם מ־`apt`.** שני מנועים, ושתי פקודות `docker` שמתחלפות לפי ה־`PATH`.
- **לדלג על בדיקת הטביעה כי „זה אתר רשמי”.** הבדיקה לא חושדת באתר; היא תופסת הורדה חלקית או קובץ אחר
  ממה שחשבת.
- **לפתוח גרסת Go חדשה מעל הישנה.** קבצים ישנים נשארים ושוברים את ההתקנה. מוחקים את התיקייה ואז פותחים.
- **להסיק מ„הפקודה עובדת” שהגרסה נכונה.** עותק ישן מוקדם ב־`PATH` עונה ראשון. `command -v` ו־`doctor.sh` אומרים
  את האמת.
- **להריץ `docker system prune` בלי לקרוא מה הוא מוחק.** הוא בטוח לנפחים, אבל `docker volume prune` לא.

## שאלות ראיון

1. **למה להריץ את התלויות של שירות (מסד נתונים, תור) בקונטיינרים גם בפיתוח מקומי?** כדי שהגרסה והתצורה יהיו
   זהות לאלה של הסביבות האחרות, ויוגדרו בקובץ שנכנס ל־Git ולא בזיכרון של מישהי. מתקינים פעם אחת את Docker
   ולא כל מסד בנפרד, ומוחקים ומקימים מחדש בשניות. המחיר: עוד שכבה להבין, צריכת זיכרון, והבדלים קטנים
   שנשארים (מערכת קבצים, רשת) בין מחשב מקומי לייצור.

2. **איך את מוודאת שקובץ הרצה שהורדת הוא מה שחשבת?** משווים את טביעת ה־SHA-256 שלו לטביעה שפורסמה. זה
   מוכיח שלמות, לא מקור: אם תוקף שולט בשרת, הוא מחליף את שניהם. למקור צריך חתימה, למשל חתימות GPG או
   Sigstore, ובארגון גם מאגר פנימי של קבצים שנבדקו. ההחלטה כמה רחוק ללכת תלויה במה שהכלי מקבל: כלי שרץ
   ב־CI עם סודות מצדיק יותר.

   **המשך:** ולמה להצמיד אימג' לפי digest ולא לפי תגית? כי תגית היא מצביע שאפשר להזיז, וה־digest מחושב
   מהתוכן. הצמדה לפי digest נותנת בנייה שחוזרת על עצמה, והמחיר הוא שעדכוני אבטחה לא מגיעים מעצמם:
   צריך תהליך שמעדכן את ההצמדות.

3. **פקודה שהתקנת „לא נמצאת”, או מדווחת גרסה אחרת. איך את מאבחנת?** קודם `command -v` לראות איזה קובץ רץ,
   ואחר כך ה־`PATH` עצמו, לפי הסדר. אחר כך: האם המעטפת נפתחה לפני ההתקנה, והאם קובץ ההגדרות באמת נטען.
   הנקודה בראיון היא השיטה: לבודד שכבה אחר שכבה במקום לנחש, ולהבין שה־`PATH` הוא רשימה מסודרת שבה הראשון
   מנצח.

4. **מה ההבדל בין טווח גרסאות בקובץ ההגדרות לבין קובץ נעילה?** הטווח אומר מה מקובל; קובץ הנעילה אומר מה
   בדיוק הותקן, כולל תלויות עקיפות וטביעות. בלי נעילה, שתי התקנות של אותו פרויקט באותו יום יכולות לקבל גרסאות
   שונות, ובאג מופיע רק אצל אחת. המחיר: מישהו צריך לעדכן את הנעילה באופן יזום, ולבדוק את העדכון.

5. **למה סקריפט בדיקה מחזיר את מספר הכשלים כקוד יציאה?** כי קוד יציאה הוא החוזה בין תוכניות: CI, Makefile
   או סקריפט אחר מחליטים לפיו בלי לפרש טקסט. אפס אומר הצלחה, וכל מספר אחר אומר כמה דברים לתקן. המגבלה:
   קודי יציאה מעל 125 שמורים למעטפת, אז סקריפט שעלול לספור יותר צריך תקרה, וקוד שמור לשגיאת שימוש (כאן 64)
   חייב להיות מתועד.

   **המשך:** ואיך בודקים סקריפט כזה בלי להתקין ולהסיר כלים? מריצים אותו מול תיקייה של כלים מזויפים, עם `PATH`
   שמכיל רק אותה, כמו `test_doctor.sh`. כך כל מקרה (חסר, ישן, חדש) הוא שורה בבדיקה.

## פתק Obsidian

הפתק המוכן נמצא גם כקובץ: `obsidian/02-workbench-wsl2.md`. זה תוכנו:

- **מערכת:** Ubuntu בתוך WSL2.
- **קונטיינרים:** Docker Desktop, על אותו מנוע.
- **כלים:** Git עם מפתח SSH, ועורך שעובד בתוך המערכת.
- **שפות:** Go ו־uv, בתיקיית הבית.
- **איפה עובדים:** רק תחת `~`. הכונן של Windows הוא לייבוא חד־פעמי.
- **איפה הכלים:** בתיקיית הבית, בלי `sudo`. מה שמוקדם ב־`PATH` מנצח.
- **הורדה בטוחה:** טביעה מאותו פרסום, בדיקה, ורק אז התקנה. טביעה מוכיחה שלמות, לא מקור.
- **הרופא:** `doctor.sh` בודק הכול; קוד היציאה הוא מספר הכשלים. ביום 57 הוספתי את חלק ה.
- **ניקוי:** מסתכלים לפני שמוחקים. נפחים נמחקים רק במפורש.
- **שאלה פתוחה:** ___

<div dir="ltr" align="left">

```text
wsl       wsl --install | wsl --list --verbose | wsl --shutdown | wsl --export Ubuntu <file.tar>
path      export PATH="$HOME/.local/go/bin:$HOME/.local/bin:$PATH"   (~/.bashrc; ~/.zshrc on a Mac)
course    export COURSE="$HOME/senior-software-engineer-course"     my work: ~/senior-course/chNN
verify    sha256sum -c <file>.sha256sum | sha256sum -c --ignore-missing <SHA256SUMS>
go        ~/.local/go  GOBIN=~/.local/bin  go version  -> go1.27.1
go tools  go install <module>@<pin>: staticcheck goose buf protoc-gen-go protoc-gen-go-grpc govulncheck actionlint
lint      golangci-lint 2.14.0: release tarball + golangci-lint-2.14.0-checksums.txt -> ~/.local/bin
uv        curl -LsSf https://astral.sh/uv/0.12.23/install.sh | sh ; uv python install 3.14
doctor    bash ch02/doctor.sh [--part-e] ; echo $?   (exit = number of failures)
docker    docker version | docker run --rm hello-world | docker system df | docker system prune
kind img  kindest/node:v1.37.0@sha256:a1ed56cf...   (pinned by digest)
my pins   go ____  uv ____  python ____  docker ____  kind ____  kubectl ____
```

</div>

## תרגיל

**מטרה:** סביבת עבודה מאומתת, וסקריפט שבודק אותה; בדיקות עשן ל־Go ול־Python; קונטיינר ראשון והתחייבות
ראשונה; וביום 57, הכלים של חלק ה וההרחבה של הסקריפט. **זמן:** שלושה ימי לימוד: יום 0 (סעיפים 2.1 עד 2.8, כ־70 דקות התקנה), יום 3 (סעיפים 2.9, 2.11 ו־2.12, כ־55 דקות בנייה) ויום 57 (סעיף 2.10, כ־45 דקות התקנה). **ידע מקדים:** פרק 0, והפרק הזה.

הצעדים המלאים, קוד ההתחלה, רשימת הבדיקה, הרמזים והפתרון נמצאים ב־
[`exercises/ch02-workbench/README.md`](../exercises/ch02-workbench/README.md).

**קריטריון סיום:**

- ביום 0, בלשונית חדשה, כל כלי מדפיס את הגרסה שמוצמדת בנספח ג, והמפתח הציבורי נוסף ל־GitHub.
- ביום 3, `doctor.sh` מסתיים בקוד יציאה 0, והקונטיינר מדפיס את הברכה.
- ביום 3, בדיקת ה־Go ובדיקות ה־pytest עוברות.
- ביום 3, ההתחייבות הראשונה מופיעה ביומן של Git, בלי `.venv` ובלי קבצי בנייה.
- ביום 57, `doctor.sh --part-e` מסתיים בקוד 0, ו־`test_doctor.sh` מדווח על 11 בדיקות שעברו.

## העשרה אופציונלית

- התיעוד של WSL, ובמיוחד העמוד על השוואת מערכות הקבצים, מסביר למה העבודה תחת `/mnt/c` איטית.
- עמוד ההתקנה הרשמי של Go מסביר את ההתקנה מהקובץ הדחוס, ואת האזהרה על פתיחה מעל גרסה קיימת.
- התיעוד של uv, בפרק על התקנה ועל ניהול גרסאות Python, מפרט מה נשמר איפה ואיך מסירים.

</div>
