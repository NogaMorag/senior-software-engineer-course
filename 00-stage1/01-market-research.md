# Stage 1 — Market Research: Senior Backend / Infrastructure Engineer, Israel, late 2026

> **Reconstructed 2026-10-05: the original stage-1 research file did not transfer from the authoring machine; this is a fresh research pass. Every claim carries a source and the date it was read.**

- **Research date:** 2026-10-05 (all sources below were read on this date unless stated otherwise).
- **Audience of the book:** an experienced developer (daily Rust, Python, older C++) preparing to be hired as a senior backend / infrastructure engineer in Israel.
- **Method:** web search plus direct fetches of job boards, recruiter salary tables, government and press reports, and engineering-hiring blogs. Job postings were collected from DevJobs (devjobs.co.il), an Israeli aggregator that republishes postings from company career pages and ATS systems. Each posting was downloaded and its text checked with a keyword script, then read by hand to check the matches.
- **Sources that blocked fetching:** Glassdoor (HTTP 403 on interview and salary pages), Indeed Israel (403), the Israel Innovation Authority's HTML press release (403; its PDF report did download and was parsed), and the Palo Alto Networks careers page (it returned the shell page, not the posting text). Where only a search-engine snippet was available, the finding says so and its confidence is lowered.
- **Confidence labels:**
  - **High:** a primary or official source, or several independent sources that agree.
  - **Medium:** one credible source, or sources that partly disagree.
  - **Low:** anecdotal, third-party, global rather than Israeli, or seen only as a snippet.

---

## 0. Executive summary

1. **"Senior" in Israeli backend/infra postings means about 5+ years.** It also means owning production end to end, leading technical design, and increasingly sharing on-call. Mentoring appears less often than you might expect. Median stated minimum in the sample: 5 years (range 3–7). *(High for the sample; the sample is small and non-random.)*
2. **Python is the most common language in Israeli senior postings (about 2 of 3).** Go appears in about a third, but usually as one item in a "one of" list. Rust appears in about 2–4%, mostly in security, endpoint and OS-level work. *(Medium.)*
3. **Kubernetes and AWS are the default infra pair.** Terraform is near-universal in DevOps/SRE postings. Kafka, Redis and PostgreSQL are named explicitly less often than you might expect (about 10–17% each). *(Medium.)*
4. **AI-assisted development is now a stated requirement in a large minority of postings.** Tools named include Cursor, Claude Code and Copilot. It is explicit in 11 of the 30 hand-checked senior postings and about 23–26% of a 178-posting sample. Building LLM/agentic features is a separate, also-growing ask. *(High that it is common; the exact share is medium.)*
5. **The Israeli interview loop still leans on a home assignment plus a team-lead / system-design interview.** Wix documents this flow publicly. Global practice is moving toward AI-allowed rounds: Meta piloted one from October 2025, Canva from June 2025. AI cheating is also pushing interviews back on-site. No Israeli company was found publicly documenting an AI-allowed coding round. *(Medium for Israel, high for the global trend.)*
6. **Salary: a senior (6–10 yrs) backend or DevOps engineer is most often quoted at about ₪35k–45k gross per month (base).** Recruiter sources range from ₪30k to ₪48k. ₪45k–55k+ goes to 9+ years, leads and AI-heavy profiles. Levels.fyi total compensation (including equity) is a median of ₪548k per year for "senior SWE, Israel". *(Medium. Sources disagree by ₪5–8k at the edges.)*
7. **The market is split.** Overall tech wages hit records (CBS average ₪36,584 in June 2026, +11.7% year on year). But software companies are cutting (6.6% layoffs in H1 2026 per an IIA survey reported by Calcalist). R&D headcount fell for the first time in a decade. Junior openings are scarce, and hiring is shifting abroad. Seniors with production, distributed-systems and AI fluency are the scarce side. *(High.)*

---

## 1. What "senior" means in 2026 backend/infra postings

### 1.1 Sample description (read carefully: small and non-random)

- **What was sampled:** 30 senior-titled postings, all Israel-located. Titles include Senior, Sr., Senior/Principal and Principal. They were collected on 2026-10-05 from DevJobs company pages for 18 employers: Check Point, Cato Networks, CrowdStrike, Dream Security, Fetcherr, NVIDIA, Palo Alto Networks, Payoneer, Cyera, JFrog, Fiverr, Wix, Lemonade, Similarweb, Taboola, Melio, Tipalti, Microsoft (Israel).
- **Dates:** 27 of the 30 postings were updated between 2026-08-25 and 2026-10-04. Three are older: Check Point/Veriti (Oct 2025, closed), Dream Posture (Mar 2026) and Cyera (Feb 2026).
- **Bias:**
  - The sample skews toward large, well-known Israeli tech and cyber companies, because those are the companies whose pages were crawled.
  - It under-represents small startups, defense, and companies hiring in Hebrew-only on AllJobs/Drushim. AllJobs and Drushim were not sampled because their listings are not easily machine-readable. This gap is stated, not hidden.
- **Lane split, by hand:** 15 backend product, 5 platform/infra-backend, 8 DevOps/SRE/platform-ops, 2 data engineering.
- **Secondary, larger sample:** 178 software-ish postings from the same crawl, updated 2026-04-04 to 2026-10-05, 84 of them senior-titled. It was keyword-counted only, not hand-read. It includes hardware-heavy employers (Apple, Intel, NVIDIA, Mobileye), which inflates C++.

All 30 posting URLs are listed in Sources [P1]–[P30].

### 1.2 Years of experience (stated minimum)

| Stated minimum | Count (n = 28 that state one) |
|---|---|
| 3 years | 2 (both Check Point "Senior") |
| 4 years | 1 |
| 5 years | 13 |
| 6 years | 6 |
| 7 years | 6 |
| not stated | 2 (Fetcherr DevOps, Tipalti) |

**Finding:** the modal and median stated minimum for "senior" is 5 years. "Principal" roles state 7+ (Palo Alto Networks Backend Infra [P14]). Payoneer's "AI Fluent" senior role asks for 7 [P15]. **Confidence: high for this sample, medium as a market statement.**

### 1.3 What the responsibilities sections actually ask for

These are quoted from the postings, read 2026-10-05.

- **Hands-on technical leadership, not management.** Palo Alto Networks (Cortex Platform): "you will be a hands-on technical leader responsible for designing, building, and operating critical backend platform services"; "Lead technical design for complex backend and platform domains, balancing long-term architecture with practical delivery" [P10].
- **Production ownership and on-call.**
  - Payoneer: "Own production from day one - monitoring, alerting, observability; you're on call for what you build" [P15].
  - Palo Alto Networks Principal Backend Infra: "The on-call rotation includes weekends and holidays", plus "Lead or contribute to customer-facing RCAs" [P14].
  - Taboola SRE: "Participate in on-call rotations, lead incident resolution and conduct blameless post-mortems" [P17].
  - Payoneer DevOps: "Available to participate in a 24/7 production on-call rotation ... is required" [P16].
- **Design docs / technical design.** Stated explicitly as "lead technical design" or "translate complex product requirements into clear technical designs and execution plans" at Palo Alto Networks [P10, P14], JFrog [P18] and Tipalti ("Lead technical design discussions and influence architectural decisions") [P29]. The words "design doc" or "RFC" did not appear in any of the 30. Postings describe the activity, not the artifact.
- **Mentoring.** Explicit in only 4 of 30: PANW Cortex "Mentor senior, mid-level, and junior engineers through design guidance, code reviews..." [P10], Tipalti [P29], CrowdStrike [P4], Similarweb Data [P25]. **Finding: mentoring is a smaller part of Israeli senior IC postings than global career ladders suggest.** Compare GitLab's public ladder, where Senior Backend Engineer explicitly includes "Provide mentorship for Junior and Intermediate Engineers" [S30].
- **End-to-end delivery.** Melio: "building a product end-to-end, through the design, coding, testing, and deployment phases" [P23]. Payoneer: "Own the full loop - understand the problem, define the solution, architect it, ship it, monitor it. No one hands you a spec" [P15].
- **Influence without authority.** PANW: "ability to influence across teams without relying on authority" [P10].

### 1.4 Keyword counts across the 30 hand-checked senior postings

Counted by regex over each posting's body (sidebar of "similar jobs" stripped). AI and Rust matches were then checked by hand. A posting counts once per item. **Small, non-random sample. Treat the numbers as directional, not as market shares.**

| Item | Postings (of 30) | Notes |
|---|---|---|
| Python | 21 | Includes "one of Python/Go/Java" lists |
| Kubernetes (incl. EKS/GKE/K8s) | 19 | |
| AWS | 18 | GCP 13, Azure 11; many list "AWS/GCP/Azure" |
| Docker / containers | 16 | |
| Go | 13 | Only ~4 make Go primary or required (Dream Posture "Go (primary language)" [P7]; CrowdStrike "new high-performance Go microservices" [P4]; Fetcherr "Go and/or Python" [P8]; Cato "Go is a plus" [P3]) |
| TypeScript / Node.js | 12 | |
| Terraform / IaC | 12 | 7 of 8 DevOps/SRE postings |
| Microservices (word) | 12 | |
| AI-assisted dev tooling (Cursor, Claude Code, Copilot, "AI-assisted", "AI tools") | 11 | Hand-verified; see §1.5 |
| CI/CD | 10 | GitHub Actions named only 2 times |
| Java | 9 | |
| On-call (explicit) | 8 | 5 of 8 SRE/DevOps, plus 3 backend/platform |
| Testing (unit/integration/TDD) | 8 | |
| MongoDB | 7 | More often than PostgreSQL (4) |
| Prometheus / Grafana | 6 | Datadog 2, OpenTelemetry 1 |
| Kafka | 5 | RabbitMQ/SQS/Pub-Sub/NATS 4 |
| Redis | 5 | |
| Helm / ArgoCD-GitOps | 5 / 5 | Almost all in DevOps/SRE/platform |
| Mentoring (explicit) | 4 | |
| C#/.NET | 4 | |
| Spark/Flink/Airflow/BigQuery etc. | 4 | |
| C++ | 3 | |
| Rust | 2 | Lemonade Platform "TypeScript proficiency with Rust experience strongly preferred" [P20]; Taboola SRE "Go, Python or Rust" [P17] |

"Observability" (incl. "monitoring") matched 22/30, but the regex is broad. A stricter count (named tools: Prometheus, Grafana, Datadog, OTel) gives 7/30. "Distributed systems" as a phrase: 14/30.

### 1.5 AI-assisted tooling: what postings literally say

These are hand-verified. 11 postings ask for AI-assisted development as a working practice:

- Similarweb Platform Foundations: "Lead the adoption and integration of AI tools to assist in code generation, testing, documentation, and debugging" [P24].
- Lemonade (Python): "Fluent with AI-powered development tools like Cursor and Claude Code" [P21].
- Payoneer: "Build with AI as your primary tool - Claude Code, Cursor ...", "Clear judgment about when to trust AI output and when to rewrite it" [P15].
- PANW AppSec: "Experience using AI-assisted engineering tools ... combined with the engineering judgment to validate generated solutions" [P11].
- Microsoft Security: "Leverage AI and AI-assisted engineering tools (such as GitHub Copilot)" [P30].
- Check Point Cloud: "Copilot, Cursor, ChatGPT, Claude, or similar" [P28].
- Tipalti: "Active, hands-on user of AI coding tools" [P29].
- JFrog: "Leverage AI tools in your daily work" [P18].
- Fetcherr DevOps: "A point of view on AI-assisted and spec-driven development" [P8].
- Check Point DevOps: "leveraging AI-powered tooling and agentic workflows to reduce toil", plus MCP integrations [P26].
- PANW Principal Backend Infra: "Incorporating AI in day to day process - CI, design, forming Agents" [P14].

Separately, about 9 postings ask for experience **building LLM/agentic product features**: Fiverr Applied AI [P19], Wix User Voice [P22], Similarweb DevOps ("LLM and agentic workflow infrastructure") [P5], Dream CTI [P6], Similarweb Data [P25], JFrog [P18], Payoneer [P15], PANW AppSec [P11], and Check Point DevOps (LLMOps) [P26].

**Finding:** "AI fluency" has moved from nice-to-have to a named qualification in about a third of senior postings in this sample. The wording consistently stresses *judgment over generated code*. **Confidence: medium-high.**

### 1.6 Larger keyword-only sample (178 postings, Apr–Oct 2026)

| Item | All 178 | Senior-titled 84 |
|---|---|---|
| Python | 53% | 63% |
| Kubernetes | 35% | 43% |
| Go | 26% | 32% |
| C++ | 30% | 24% (hardware/embedded employers) |
| TypeScript/Node | 23% | 26% |
| Java | 24% | 23% |
| AI dev tools | 23% | 26% |
| Terraform | 15% | 18% |
| Kafka | 10% | 13% |
| Redis | 10% | 10% |
| PostgreSQL | 10% | 7% |
| C#/.NET | 11% | 10% |
| Rust | 4% (8 postings) | 2% (2) |

The 8 Rust mentions are:
- Cato Networks Endpoint DLP
- Lemonade Platform
- three Check Point roles (incl. "Container Network Security", "Workforce AI Security")
- Microsoft Windows Silicon integration
- Cato AI Security team lead
- Apple iOS

Source: DevJobs pages crawled 2026-10-05 [S1]. **Confidence: medium.** It is keyword-only, and the employer mix is biased toward big companies and hardware.

---

## 2. Lanes and how they differ in postings

Based on the 30-posting sample (§1). Counts are within each lane.

| Lane | n | What dominates | Distinguishing asks |
|---|---|---|---|
| **Backend product** | 15 | Python 9, TypeScript/Node 7, Java 6, Go 5. Microservices 9. MongoDB 5. Kafka 3 | Feature/domain ownership end to end, APIs, data modelling, working with Product. On-call explicit in only 1. AI-tooling 6. Several want LLM product features (Fiverr, Wix, Payoneer). |
| **Platform / infra backend** | 5 | Kubernetes 5/5, AWS/GCP 4, Python 4, Go 3 | Building shared internal services ("internal platforms, shared services, developer platforms" [P10]), multi-tenant systems, reliability for internal customers, on-call 2/5, K8s operators (JFrog [P18]). |
| **DevOps / SRE / platform-ops** | 8 | Kubernetes 7, Terraform 7, Docker 7, Python 7, Go 5, Prometheus/Grafana 5, ArgoCD/GitOps 4, Helm 3 | On-call 5/8, incident response and post-mortems, IaC, CI/CD, observability stacks, multi-cloud, cost. Coding expected for tooling ("Build internal software tooling ... in Go, Python or Rust" [P17]). |
| **Data infrastructure** | 2 | Spark/Flink/Airflow, AWS | Pipelines, data collection at scale. Sample too small to generalise; GotFriends separately reports Airflow-proficient data engineers averaging ₪44.5k [S12]. |

**Observations:**
- The line between "senior backend" and "platform engineer" is blurry. Platform-flavoured backend roles still read as software roles (design, APIs, data modelling) with Kubernetes added. DevOps/SRE roles read as operations roles with coding added.
- An engineer coming from Rust plus systems work is closest to the **platform/infra-backend** lane and the **security-product backend** lane (cyber is heavily represented in Israel).

**Confidence: medium** (small n per lane).

---

## 3. The 2026 senior interview loop

### 3.1 Israel

- **Wix (official careers page) documents a 6-stage flow** [S20]:
  1. Apply.
  2. First professional interview ("a short call or a 1-hour interview").
  3. **Home assignment** ("a short assignment to complete at home", which Wix calls "the best equalizer").
  4. Second professional interview with the team lead, discussing the assignment.
  5. HR interview(s).
  6. Offer, with the aim to finish "in one month".

  Wix's public server-guild repo (Lithuania office) describes a homework assignment of "1-2 evenings" ("You pick your own technology and deadline"). It is followed by a 2–3 hour technical interview focused on system design, where "Problems you are given don't have objectively correct answers" [S21]. **Confidence: high for Wix; medium as representative of Israel.**
- **monday.com (Glassdoor, seen only as a search snippet; the page returned 403):** a Tel Aviv backend candidate reported three technical interviews. One was coding/problem solving and two were system design, one of them "designing ... the notification system". Two HR interviews followed [S22]. **Confidence: low** (single anonymous report, snippet only).
- **A recruiter guide (Nisha, 2025-06-13)** lists the typical Israeli developer loop [S23]:
  - ATS keyword screening.
  - Live coding of 45–90 minutes.
  - A 48-hour take-home with tests (">80% coverage"), a README and CI (GitHub Actions).
  - Pair programming plus system design ("bottlenecks and costs").
  - Behavioral questions in STAR form, and negotiation using Levels.fyi data.

  Its headline: "Today we check how you think in Production, from understanding strategy to measured rollback". **Confidence: medium** (recruiter marketing content, but consistent with the other sources).
- **Home-assignment review criteria (Geektime, 2018-10-21, old but still cited):** reviewers check that the code runs, style/linters, tests, the README, and that it uses Git with logical commits. "Your code sample testifies to a developer candidate no less than their CV" [S24]. **Confidence: medium** (dated).
- **AI in the Israeli process:**
  - Fiverr piloted *recorded* async video interviews in place of phone screens, with an AI-generated analysis for recruiters ("The decisions are human-made, but AI is involved in the process"; Geektime 2026-02-19) [S25]. This is AI used *by the employer*, not AI allowed for the candidate.
  - Bizportal (2026-08-06) reports that AI-enabled fraud is pushing interviews back to in-person [S26]. Google and Cisco are cited; McKinsey bans AI and requires visible hands. It also cites a Gartner survey: 6% of 3,000 candidates admitted to interview fraud.
  - **No Israeli company was found publicly documenting an AI-allowed live coding round. Not found ≠ does not exist.**
  - **Confidence: medium.**

### 3.2 Global (shapes what multinationals in Israel do)

- **Meta AI-enabled coding round:**
  - Piloted from October 2025.
  - 60 minutes in CoderPad with a built-in assistant (GPT-5, Claude Sonnet 4/4.5, Gemini 2.5 Pro, Llama 4 and others).
  - "a multi-file project that you have to iterate on instead of two algorithmic problems".
  - Expected to roll out "for all back-end and ops-focused roles in 2026" (interviewing.io, Oct 2025) [S27]. Meta's other coding round remains no-AI.
  - **Confidence: high** that it exists; medium on rollout breadth.
- **Canva (engineering blog, 2025-06-11):**
  - Backend/ML/frontend candidates are *expected* to use Copilot/Cursor/Claude.
  - Realistic problems ("Build a control system for managing aircraft takeoffs and landings").
  - Evaluated on breaking down ambiguity, finding and fixing issues in AI-generated code, and production standards [S28].
  - **Confidence: high.**
- **interviewing.io survey (Oct 2025, 67 interviewers, 52 at FAANG)** [S29]:
  - 0 of 52 said their company dropped algorithmic questions.
  - 81% suspected AI cheating; about a third caught it.
  - 58% changed their questions toward custom, multi-part problems.
  - 50% expect a return to in-person.
  - 67% of startup respondents reported meaningful changes, "including eliminating take-homes and adding AI-assisted interviews".
  - **Confidence: high** (primary; small survey, US-centric).
- **Take-home decline:** one 2026 blog cites a Karat survey (400 engineering leaders, "71% say AI is making technical skills harder to assess"). It also claims 40–60% senior drop-off on long take-homes, with no attributed source [S31]. **Confidence: low.**
- **Rippling / Shopify AI-allowed rounds:** reported only by interview-prep vendors [S32]. **Confidence: low.**
- **Practical coding questions** (rate limiter single-node vs distributed, LRU cache with TTL, top-K on a stream, task scheduler) are commonly cited as backend-flavoured coding rounds. The sources are prep sites and question banks [S33]; no Israeli primary source was found. **Confidence: low-medium.**
- **Code-review round** (review a seeded broken PR in about 30 minutes, scored on ranking issues by severity): the evidence is a practitioner blog post (dev.to) [S34]. **Confidence: low.** Plausible and growing, but thinly evidenced.

### 3.3 Synthesis: what to prepare for in Israel, late 2026

1. Screen call with the team lead or a recruiter (CV deep-dive).
2. **Either** a home assignment (1–2 evenings, still common in Israel) **or** live coding of 45–90 minutes (practical rather than LeetCode-hard).
3. One or two **system design** rounds, often a domain of the company's own product (notifications, multi-tenant SaaS, data pipelines).
4. A "project drill-down" or past-project architecture discussion.
5. HR/behavioral rounds.

Multinationals with Israeli R&D (Meta, Microsoft, Google, etc.) increasingly add AI-allowed rounds, while also moving rounds back on-site. **Confidence: medium.**

---

## 4. Go vs Python vs Rust demand in Israel

- **Python leads by a wide margin** in Israeli backend/infra postings: 21/30 senior hand-checked postings and 53–63% of the 178-posting sample (§1). Globally, Python is used by 54.8% of professional developers (Stack Overflow 2025) [S35]. **Confidence: high.**
- **Go is a solid second tier for backend/infra/DevOps:** 13/30 and 26–32% of the larger sample. But it is usually one item in "Go, Python, Java or similar". Few postings make Go primary or required (see §1.4). GotFriends lists a separate "Go Developer" salary row [S12], which shows it is a recognised hiring category. Globally, 17.4% of professional developers use Go [S35]. **Confidence: medium.**
- **Rust demand in Israel is real but niche:** 2/30 senior hand-checked and 8/178 (4%) in the larger sample. Globally, 14.5% of professional developers use Rust [S35], so Israeli *postings* lag Rust *usage*. **Confidence: medium.** Where it exists:
  - **Security / endpoint / network security:**
    - Check Point Container Network Security and Workforce AI Security, Cato Endpoint DLP (from the crawl [S1]).
    - Palo Alto Networks "Senior Rust Software Engineer - Data Classification (Cortex Cloud)" (search result only; the page did not render) [S36].
    - The rust.org.il company list adds Varonis (endpoint DLP agent), Intezer, Spectral and Cynerio [S37].
  - **Crypto / Web3:** web3.career lists 18 Rust-tagged Web3 jobs in Tel Aviv (Chainalysis, Fireblocks and others) [S38]. StarkWare and Fireblocks are on the rust.org.il list [S37]. **Confidence: medium.**
  - **Infra / data platforms:** Coralogix, Pinecone, Redis and Pelanor are listed as Rust users [S37]. Lemonade Platform prefers Rust [P20].
  - **Fintech / algorithmic trading:** a stealth Kfar Saba company "building an algorithmic trading infrastructure in Rust" (rust.org.il jobs page) [S39].
  - **Embedded / IoT:** Greenvibe and Nanit are listed [S37]. Apple and Microsoft OS-level roles mention Rust [S1].
  - **Dev tools:** Tabnine (client side), MetalBear (mirrord) [S37].
  - **Caveat:** the rust.org.il list is community-maintained and undated. Being listed means "uses Rust somewhere", not "is hiring Rust engineers now".
- **Salary premium for Go/Rust: contradictory.**
  - Digital Hunters claims "Go, Rust, and Kotlin-heavy backgrounds tend to command the top of the range" [S10].
  - GotFriends' own September 2026 table shows Go 6–10 yrs at ₪33–45k vs Python ₪34–45k, i.e. *no* Go premium [S12].
  - **Confidence: low** that any language premium exists. Lane and domain (security, infra, AI) matter more than language.

**Implication:**
- Rust alone narrows the Israeli target market to security/endpoint, crypto, a few infra/data-platform companies and HFT.
- Python (senior-level) plus Go opens most of the backend/platform/DevOps market.
- Rust is a strong *differentiator* for systems-heavy roles, not a primary entry ticket.

---

## 5. Salaries: senior backend / infra, Israel 2025–2026 (gross monthly NIS)

All figures are gross monthly salary in ₪ unless noted. **Recruiter tables are not random samples.** A recruiter's pool skews upward, a point Scalefox itself makes [S11].

| Source (date read 2026-10-05) | Data date | Role / band | ₪ / month |
|---|---|---|---|
| GotFriends salary tables [S12] | "Last updated September 2026" | Backend Developer 6–10 yrs | 35–45k |
| GotFriends [S12] | Sep 2026 | Backend, management tier | 42–55k |
| GotFriends [S12] | Sep 2026 | Go 6–10 yrs / Python 6–10 yrs / Java 6–10 yrs / Node 6–10 yrs | 33–45k / 34–45k / 38–45k / 33–44k |
| GotFriends [S12] | Sep 2026 | DevOps 6–10 yrs / Data Engineer 6–10 yrs | 35–45k / 33–43k |
| GotFriends [S12] | H1 2026 | Average for candidates with 5+ yrs | 41,722 (from 39,810 in 2025) |
| Scalefox / JobFox blog [S11] | 2026-07-09 (based on GotFriends pool + Ravio 2026 + CBS) | Backend senior 6–10 yrs / DevOps-SRE senior / Data-ML senior | 38–48k / 38–48k / 40–55k |
| Digital Hunters [S10] | 2026 article, "2024-2025 market data" from own placements | Senior Backend 6+ / DevOps-SRE 6+ / Data Eng 6+ (Tel Aviv) | 30–42k / 28–40k / 26–38k |
| Ethosia salary tables [S13] | **Undated page** | Java/Python/C#/C++ developer 5–8 yrs / 9+ yrs | 35–45k / 45–55k |
| Aman Group via Maariv [S14] | 2026-07-26 | Senior developers / tech leads / architects | 30–40k / 32–42k / 35–45k |
| GotFriends via JPost [S15] | 2025 data | Average tech salary 2025; LLM/RAG/NLP specialists; team leads/dev managers | 39,810; 43,212; 56,646 |
| CBS via JPost [S16] | June 2026 | Average wage, tech sector (all roles, all levels) | 36,584 (+11.7% YoY) |
| Levels.fyi [S17] | page "last updated" 2026-10-05; 967 submissions | Senior SWE Israel, **total comp per year** (base + bonus + equity) | median ₪548,432/yr (p25 ₪461k, p75 ₪674k, p90 ₪865k) ≈ ₪45.7k/month TC-equivalent |
| Glassdoor | — | — | **Blocked (403); not used** |
| SeeV | — | — | **Not found** in this pass |

**Synthesis:**
- **Central band for a senior (about 6–10 yrs) backend or infra IC: about ₪35k–45k gross monthly base.** Most sources overlap there. **Confidence: medium.**
- **Lower edge:** Digital Hunters (₪30k, based on older 2024–25 data) and Aman (₪30–40k "senior developers").
- **Upper edge:** Scalefox (₪48k, stated upward skew) and Ethosia's 9+ band (₪45–55k).
- **Dispersion drivers** cited in the sources:
  - Location: 10–20% Tel Aviv premium vs Haifa/Be'er Sheva [S10].
  - Domain: security research about ₪57.6k [S12], vulnerability-research outliers about ₪120k [S15], AI/RAG ₪45–46k [S15].
  - Equity at multinationals: Levels.fyi's p75–p90 total comp is far above recruiter base tables.
  - Defense-tech is closing the gap with SaaS and cyber (GotFriends, JPost 2026-08-01) [S18].
- **Uncertainty:**
  - Recruiter tables don't publish methodology or sample sizes per cell.
  - Levels.fyi skews toward large multinationals and total comp.
  - The strong-shekel pressure on dollar-revenue companies (Maariv [S14]) may cap 2027 raises.
- **Never quote a single number to the reader.** Give the band plus the drivers.

---

## 6. Market conditions, late 2025 – 2026

- **Wages at records, but headcount flat or shrinking in software.**
  - CBS: average tech wage ₪36,584 in June 2026, +11.7% YoY. The tech workforce was stable at about 396,900 jobs (JPost, 2026-09-06) [S16]. **High.**
  - Israel Innovation Authority 2026 report (via Calcalist, 2026-05-31) [S19]:
    - R&D employment fell by about 3,500, "the first decline in a decade".
    - Product-related roles grew by about 15,000.
    - The share of Israeli employees in private tech companies fell from 69% (Jan 2019) to 62% (Mar 2026).
    - Wix (1,000), Meta Israel (100+) and Amdocs (2,500–3,000 globally) layoffs were reported.

    **High.**
- **AI and layoffs (IIA employer survey, reported by Calcalist 2026-08-11)** [S5]:
  - 210 companies with about 130,000 employees.
  - Software-company layoffs 6.6% in H1 2026 vs hardware 1.1%.
  - Over a third expect employment declines in H2 2026.
  - Companies that reduced hiring after adopting AI tripled from 3% to 10%.
  - Widespread AI adoption rose from 21% to 30% in six months.

  **High.**
- **Earlier IIA snapshot (Jan 2026 PDF, 500 employers)** [S6]:
  - 35% of tech companies laid off staff in the prior six months (vs 22% in other sectors).
  - Only 5% named AI as the key reason.
  - 94% recruited during the year; 23% expect to recruit less in 2026.

  This shows the AI effect *accelerating* between the January and August surveys. **High.**
- **Juniors hit hardest.**
  - Ethosia data (Globes, 2025-06-11): only 88 of 5,641 open tech positions (1.5%) were for candidates with no experience. For 1–3 years: about 500 positions vs about 1,400 seekers [S7]. **High** (named data source).
  - Calcalist (2025-07-14): tech job seekers rose 112% from 2019 to April 2025. Openings per tech job seeker fell from 1.6 (2019) to 0.9 (2024–25). Higher-salary earners rose from 15% to 40% of tech job seekers [S8]. That last figure means seniors are not immune. **High.**
  - Aman Group (Ynet, 2026-01-06): "Demand for junior professionals is declining". Hiring cycles grew from about 30 days to 45–60 days. Hiring outside Israel rose more than 23% [S9]. **Medium** (a recruiter's own data, about 440 positions).
  - Counter-trend: Calcalist (2026-05-01) reports Salesforce, IBM and Dropbox increasing entry-level hiring globally. It also says only 10% of Israel's 17,000 open tech roles were entry-level [S3]. **Medium.**
  - Global: Stanford Digital Economy Lab (Brynjolfsson, Chandar, Chen; revised 2026-08-12) finds employment of 22–25-year-olds in AI-exposed occupations is "19% below where it would be", mainly through reduced hiring [S2]. **High** (US data, global relevance).
- **Seniors with the right profile are scarce.**
  - Digital Hunters lists "senior backend engineers with distributed systems expertise (Kafka, Kubernetes, multi-region architecture)" among three shortage areas. Senior backend time-to-hire is 6–10 weeks [S4]. **Low-medium** (no external sources cited).
  - GotFriends (JPost) says companies showed "a clear preference for senior candidates with at least five years of experience, particularly those with deep mastery of AI tools" [S15]. **Medium.**
- **Bright spots:**
  - Defense-tech new jobs +20% in 2026, wages +8% (GotFriends, JPost 2026-08-01) [S18].
  - Hardware/chips expanding while software contracts [S5].
  - Hot fields per Aman: Data & AI, Cloud, DevOps, Cybersecurity [S9].

**Net reading:**
- For an experienced engineer, late 2026 is a *selective* market, not a closed one.
- Fewer seats in pure-software R&D, longer hiring cycles, and more on-site and anti-cheating scrutiny.
- A premium for production ownership, cloud-native infra and AI-assisted productivity.
- Security, infra and defense are relatively strong. **Confidence: medium-high.**

---

## 7. Implications for the book

Priorities are derived from the evidence above. The section references show where each came from.

| Area | Priority | Why (evidence) | What the book should emphasise |
|---|---|---|---|
| **Go** | High | 13/30 senior postings, 26–32% of the larger sample, strong in platform/DevOps (§1.4, §4) | Idiomatic Go services: context, cancellation, goroutine/channel patterns, error wrapping, `net/http`, gRPC, testing, profiling (pprof). Frame Go as Rust-to-Go translation (ownership vs GC, interfaces vs traits). Get to "can pass a practical Go coding round". |
| **Python (senior)** | Highest | 21/30 and about 63% of senior postings (§1.4, §1.6) | Senior-level Python, not scripting: typing (mypy/pyright), asyncio, FastAPI, packaging (uv), pytest, performance limits (GIL, multiprocessing), production patterns. |
| **API / DB / Kafka / Redis** | High | Microservices 12/30, MongoDB 7, PostgreSQL 4 (+10% wider), Kafka/Redis 5 each (§1.4); Digital Hunters shortage list (§6) | REST + gRPC API design, idempotency, pagination, versioning. PostgreSQL (indexes, transactions, isolation, migrations) plus a document-store contrast. Kafka semantics (partitions, consumer groups, delivery guarantees, outbox). Redis (caching patterns, rate limiting, locks and their pitfalls). |
| **Distributed systems** | High | 14/30 phrase mentions; shortage area (§1.4, §6); system design rounds (§3) | Consistency models, replication, partitioning, retries/backoff/timeouts, idempotency, backpressure, leader election basics, multi-region trade-offs. Always tie to production failure stories. |
| **Docker / Kubernetes / Terraform / GitHub Actions** | High (K8s), Medium-high (Terraform), Medium (GHA) | K8s 19/30; Terraform 12/30 (7/8 SRE); CI/CD 10, GHA named 2 (§1.4) | Containers done properly (multi-stage, non-root), Kubernetes primitives and debugging, Helm/ArgoCD basics, Terraform modules/state, a CI pipeline in GitHub Actions as the worked example (portable concepts). |
| **Observability (OTel / Prometheus / Grafana)** | Medium-high | On-call 8/30; Prometheus/Grafana 6, OTel 1, Datadog 2 named (§1.4); "own production from day one" [P15] | Teach the *practice* (SLIs/SLOs, RED/USE, alert design, tracing, incident response, blameless post-mortems) with OTel + Prometheus + Grafana as the open stack. Note that vendors (Datadog, Coralogix) are common in Israel. |
| **Testing & security** | Medium-high | Testing explicit 8/30; take-home rubrics stress tests/README/CI (§3.1); security-product companies dominate Israeli hiring (§1.1, §4) | Test pyramid, contract/integration tests, testcontainers, property testing. Secure-by-default backend (authn/z, secrets, OWASP API Top 10, supply chain). Leverage the Rust/C++ background for security-domain roles. |
| **AI-assisted engineering** | High (cross-cutting) | 11/30 explicit; about 9/30 want LLM-feature experience; AI-allowed rounds globally (§1.5, §3.2) | Using Claude Code/Cursor/Copilot *with judgment*: review generated code, write specs and tests first, recognise subtle bugs. Rehearse an AI-allowed coding round (Meta/Canva format) and a no-AI round. Basic LLM-feature engineering (RAG, evals, cost/latency, prompt injection) as a stretch module. |
| **System design** | Highest | Every Israeli loop found includes it, often 2 rounds and product-domain (§3.1) | A repeatable framework, 10–15 worked designs (notification system, rate limiter, job scheduler, multi-tenant SaaS, event pipeline, URL/feature-flag service). Design-doc writing (postings ask for "technical designs and execution plans"). Cost and operability as first-class concerns. |
| **Practical coding round** | High | Live coding 45–90 min or home assignment (§3.1); rate limiter / LRU / scheduler style (§3.2, low confidence) | Practice in Go and Python. Home-assignment hygiene (README, tests, CI, Git history). Code-review round practice (low-confidence trend, cheap to prepare). |
| **Career** | High | 5+ years is the bar; seniors scarce but market selective; salary dispersion (§1.2, §5, §6) | Positioning a Rust/systems background for Israeli lanes (security, platform, infra). CV keywords (ATS screening [S23]). Project drill-down story prep. Behavioral (STAR) with ownership/incident stories. Salary negotiation with bands, not single numbers. Target-company mapping (cyber, infra, defense-tech, multinationals). |

**Book-level recommendations:**
1. Make **Python-senior + Go** the two working languages. Use Rust as the bridge and differentiator, not the main vehicle.
2. Treat **system design + production ownership** as the core thread through the whole book. That is what "senior" means in the postings.
3. Treat **AI-assisted engineering as a skill to be examined**, not a shortcut. Teach verification habits.
4. Keep salary and market chapters **dated and re-checkable**. Every number in this file has a date for that reason.

---

## 8. Known gaps in this research

- AllJobs, Drushim, LinkedIn and Comeet were not systematically sampled. All postings came via DevJobs company pages for 18 mostly large employers (30-posting sample).
- Glassdoor and Indeed were blocked, so there is no Glassdoor salary or interview-review data beyond one snippet.
- SeeV salary data: not found.
- No Israeli primary source on AI-allowed candidate interviews was found.
- Hebrew forum discussions (e.g. Facebook groups, Reddit r/israel tech threads) were not mined. Interview-loop findings for Israel rest mainly on Wix, a recruiter guide, Geektime/Bizportal and one Glassdoor snippet.
- Keyword counts are regex-based. "Observability" and "design" matches are broad. AI and Rust matches were hand-checked; the other matches were not exhaustively hand-checked.

---

## Sources

All read on 2026-10-05.

### Job postings (30-posting hand-checked sample; DevJobs mirrors of employer postings)

- [P1] Check Point (Veriti) — Senior Backend Engineer, Tel Aviv (updated 2025-10-23, closed). https://www.devjobs.co.il/job-details/4319073644
- [P2] Dream Security — Senior Backend Engineer, Posture (2026-03-09). https://www.devjobs.co.il/job-details/4340340593
- [P3] Cato Networks — Senior Software Engineer (2026-09-24). https://www.devjobs.co.il/job-details/4441520093
- [P4] CrowdStrike — Sr. Software Engineer, Cloud Identity Protection (2026-09-30). https://www.devjobs.co.il/job-details/4442890703
- [P5] Similarweb — Senior DevOps Engineer (2026-09-24). https://www.devjobs.co.il/job-details/4416097902
- [P6] Dream Security — Senior Backend Engineer, CTI (2026-08-29). https://www.devjobs.co.il/job-details/4434986281
- [P7] (= P2, Go "primary language" quote). https://www.devjobs.co.il/job-details/4340340593
- [P8] Fetcherr — Senior DevOps Engineer, Netanya (2026-09-17). https://www.devjobs.co.il/job-details/4429844723
- [P9] NVIDIA — Senior Software Engineer, Networking, Raanana (2026-10-03). https://www.devjobs.co.il/job-details/4464816368
- [P10] Palo Alto Networks — Senior/Principal Backend Engineer, Cortex Platform (2026-10-03). https://www.devjobs.co.il/job-details/4435690210
- [P11] Palo Alto Networks — Senior/Principal Backend Engineer, AppSec Platform (2026-10-02). https://www.devjobs.co.il/job-details/4454145375
- [P12] Palo Alto Networks — Senior/Principal SRE (2026-09-27). https://www.devjobs.co.il/job-details/4470801230
- [P13] Cyera — Senior DevOps Engineer (2026-02-01). https://www.devjobs.co.il/job-details/4367998817
- [P14] Palo Alto Networks — Principal Engineer Software, Backend Infra (2026-09-24). https://www.devjobs.co.il/job-details/4469753691
- [P15] Payoneer — Senior Software Engineer (AI Fluent), Herzliya (2026-09-23). https://www.devjobs.co.il/job-details/4449648811
- [P16] Payoneer — Senior DevOps Engineer, Herzliya (2026-08-25). https://www.devjobs.co.il/job-details/4458787797
- [P17] Taboola — Senior SRE (2026-09-28). https://www.devjobs.co.il/job-details/4445605253
- [P18] JFrog — Senior Software Engineer (2026-09-27). https://www.devjobs.co.il/job-details/4471093163
- [P19] Fiverr — Senior Backend Engineer, Applied AI (2026-09-23). https://www.devjobs.co.il/job-details/4469452841
- [P20] Lemonade — Senior Backend Engineer, Platform (2026-10-01). https://www.devjobs.co.il/job-details/4440577997
- [P21] Lemonade — Senior Backend Engineer, Python (2026-10-01). https://www.devjobs.co.il/job-details/4440845950
- [P22] Wix — Senior xEngineer (Backend Oriented), User Voice (2026-09-08). https://www.devjobs.co.il/job-details/4461000093
- [P23] Melio — Senior Backend Engineer (2026-09-25). https://www.devjobs.co.il/job-details/4443861537
- [P24] Similarweb — Senior Backend Engineer, Platform Foundations (2026-10-04). https://www.devjobs.co.il/job-details/4420292899
- [P25] Similarweb — Senior Data Engineer (2026-10-01) and Senior Data Engineer, Data Collection (2026-09-16). https://www.devjobs.co.il/job-details/4456700791 ; https://www.devjobs.co.il/job-details/4412573277
- [P26] Check Point — Senior DevOps Engineer (2026-09-07). https://www.devjobs.co.il/job-details/4464317871
- [P27] Dream Security — Senior DevOps Platform Engineer, CTI (2026-09-14). https://www.devjobs.co.il/job-details/4465727084
- [P28] Check Point — Senior Cloud Software Developer (2026-10-04). https://www.devjobs.co.il/job-details/4475374996
- [P29] Tipalti — Senior Software Engineer (2026-09-23). https://www.devjobs.co.il/job-details/4469318663
- [P30] Microsoft (Israel) — Backend Software Engineer II / Senior, Microsoft Security (2026-10-02). https://www.devjobs.co.il/job-details/4465686053

(P1–P30 cover 30 distinct postings. P7 repeats P2 for a quote, and P25 covers two Similarweb data postings.)

### Other sources

- [S1] DevJobs company job grids (crawl of 33 employer pages; 178 postings updated 2026-04-04..2026-10-05). https://www.devjobs.co.il/jobs-grid?companyId=palo-alto-networks (pattern `?companyId=<slug>`)
- [S2] Stanford Digital Economy Lab — Brynjolfsson, Chandar, Chen, "Canaries in the Coal Mine" (revised 2026-08-12). https://digitaleconomy.stanford.edu/publications/canaries-in-the-coal-mine/
- [S3] Calcalist — "Juniors are back?" (2026-05-01). https://www.calcalist.co.il/calcalistech/article/bjphxyxc11g
- [S4] Digital Hunters — Israel Tech Labor Market 2026 (2026-05-07). https://dhh.agency/2026/05/07/israel-tech-labor-market-2026-state-of-hiring-salaries-and-talent-supply/
- [S5] Calcalist Tech — "AI is splitting Israel's high-tech industry in two" (2026-08-11). https://www.calcalistech.com/ctechnews/article/bkn11abdlmx
- [S6] Israel Innovation Authority — "Snapshot: The Effect of AI on Employment in the High-Tech Industry" (PDF, Jan 2026). https://innovationisrael.org.il/en/wp-content/uploads/sites/3/2026/01/Innovation-Authority-Zviran-Employers-Survey-Report-ENG-29_1_2026-.pdf
- [S7] Globes — junior crisis, Ethosia data (2025-06-11). https://www.globes.co.il/news/article.aspx?did=1001512461
- [S8] Calcalist Tech — special report on the employment crisis (2025-07-14). https://www.calcalistech.com/ctechnews/article/bjv7kvglxe
- [S9] Ynet News — "Crisis or opportunity? What the 2026 technology job market is really telling us" (2026-01-06). https://www.ynetnews.com/business/article/skofzljxze
- [S10] Digital Hunters — Tech Salaries in Israel 2026 (2026-05-07). https://dhh.agency/2026/05/07/tech-salaries-in-israel-2026-role-by-role-benchmarks-for-hiring-managers/
- [S11] Scalefox/JobFox — Israel Tech Salaries 2026 (2026-07-09). https://jobs.scalefox.ai/blog/en/israel-tech-salaries-2026/
- [S12] GotFriends — High-tech salary tables 2026 (last updated Sep 2026). https://www.gotfriends.co.il/%D7%91%D7%9C%D7%95%D7%92%D7%99%D7%9D/high-tech-salary-tables/
- [S13] Ethosia — salary tables (undated page). https://ethosia.co.il/%D7%98%D7%91%D7%9C%D7%90%D7%95%D7%AA-%D7%A9%D7%9B%D7%A8/
- [S14] Maariv — juniors and QA out, hiring moves to Romania/Poland (Aman Group, 2026-07-26). https://www.maariv.co.il/economy/tech/article-1348750
- [S15] Jerusalem Post — "AI boom pushes Israeli tech salaries toward NIS 40k in 2026" (GotFriends 2025 data). https://www.jpost.com/business-and-innovation/article-882590
- [S16] Jerusalem Post — CBS average wage, June 2026 (2026-09-06). http://www.jpost.com/business-and-innovation/banking-and-finance/article-907682
- [S17] Levels.fyi — Senior Software Engineer, Israel. https://www.levels.fyi/en-gb/t/software-engineer/levels/senior/locations/israel
- [S18] Jerusalem Post — defense-tech hiring +20% (2026-08-01). https://www.jpost.com/defense-and-tech/article-904276
- [S19] Calcalist Tech — "Israel's startups grow, but not necessarily in Israel" (IIA 2026 report, 2026-05-31). https://www.calcalistech.com/ctechnews/article/sk8qqitegg
- [S20] Wix Careers — How We Hire. https://careers.wix.com/how-we-hire
- [S21] Wix Incubator — Server Guild (LT) interview process (GitHub, undated). https://github.com/wix-incubator/server-guild-lt-interviews/blob/master/PROCESS.md
- [S22] Glassdoor — monday.com Backend Developer interviews (search snippet only; page 403). https://www.glassdoor.com/Interview/monday-com-Backend-Developer-Interview-Questions-EI_IE725019.0,10_KO11,28.htm
- [S23] Nisha — "What does a programmer job interview look like in 2025?" (2025-06-13). https://www.nisha.co.il/%D7%90%D7%99%D7%9A-%D7%A0%D7%A8%D7%90%D7%94-%D7%A8%D7%90%D7%99%D7%95%D7%9F-%D7%A2%D7%91%D7%95%D7%93%D7%94-%D7%9E%D7%AA%D7%9B%D7%A0%D7%AA/
- [S24] Geektime — what reviewers look for in developer home assignments (2018-10-21). https://www.geektime.co.il/dev-homework-home-test/
- [S25] Geektime — Fiverr adds AI to candidate vetting (2026-02-19). https://www.geektime.co.il/fiverr-adds-ai-to-employee-vetting-process/
- [S26] Bizportal — AI is bringing job interviews back to the room (2026-08-06). https://www.bizportal.co.il/career/news/article/20038103
- [S27] interviewing.io — How to use AI in Meta's AI-assisted coding interview (Oct 2025). https://interviewing.io/blog/how-to-use-ai-in-meta-s-ai-assisted-coding-interview-with-real-prompts-and-examples
- [S28] Canva Engineering — "Yes, You Can Use AI in Our Interviews. In fact, we insist" (2025-06-11). https://www.canva.dev/blog/engineering/yes-you-can-use-ai-in-our-interviews/
- [S29] interviewing.io — "How is AI changing interview processes? Not much and a whole lot" (Oct 2025). https://interviewing.io/blog/how-is-ai-changing-interview-processes-not-much-and-a-whole-lot
- [S30] GitLab Handbook — Backend Engineer job family. https://handbook.gitlab.com/job-families/engineering/development/backend/
- [S31] JobsByCulture — Take-Home vs Live Coding in 2026 (2026-06-05). https://jobsbyculture.com/blog/take-home-vs-live-coding-2026
- [S32] Final Round AI — Companies that allow AI during interviews (2026; vendor content). https://www.finalroundai.com/blog/companies-that-allow-ai-during-interviews
- [S33] interviewing.io — LRU Cache question page; plus prep-site search results on rate-limiter/LRU/top-K backend rounds. https://interviewing.io/questions/lru-cache
- [S34] DEV Community — "The Code Review Interview: I Hand Candidates a Broken PR". https://dev.to/ji_ai/the-code-review-interview-i-hand-candidates-a-broken-pr-3mam
- [S35] Stack Overflow Developer Survey 2025 — Technology. https://survey.stackoverflow.co/2025/technology
- [S36] Palo Alto Networks — Senior Rust Software Engineer, Data Classification (Cortex Cloud) (search result; page content did not render). https://jobs.paloaltonetworks.com/en/job/tel-aviv/senior-rust-software-engineer-data-classification-cortex-cloud/47263/76996242160
- [S37] Rust Israel community — companies using Rust in Israel (undated). https://rust.org.il/companies/
- [S38] web3.career — Rust Web3 jobs in Tel Aviv. https://web3.career/web3-jobs-tel-aviv+rust
- [S39] Rust Israel community — jobs page. https://rust.org.il/jobs/
