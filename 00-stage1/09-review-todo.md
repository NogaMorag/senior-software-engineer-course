# Stage 3 to-do (collected during Stage 2)

- Language: chapters 9, 27, 28 and `exercises/ch28-security/README.md` render tenant as דייר; convert to דיירת with agreement (TERMINOLOGY.md).
- Language: author references "הכותב" in chapters 19, 20, 25 (use "בהרצה שבה נכתב הספר").
- Appendix references "נספח C" in chapter 20 (and any later chapter): use "נספח ג".
- Chapter 2 is 82 KB: consider moving per-tool troubleshooting rows to appendix B.
- golangci-lint v2.14.0 is installed in the ch27 README, not in chapter 2 §2.7.1: move it to Day 0 and doctor.sh.
- Checker: "פרקים 29 ו־36" form of chapter references is not parsed (extend the regex).
- Chapter 34: verify notice-period / pension / keren hishtalmut statements on primary sources (kolzchut returned 403 to the writer).
- Chapter 28: RFC 8725, 7636, 9700 numbers stated from memory; verify on rfc-editor.org.
- Chapters 5-6: Rust/tokio box claims not checked against primary docs.
- Appendix C: build from 07-tech-stack.md incl. the writer-added pins.
- Chapter 2 §2.12 promises appendix B covers "WSL virtual disk does not shrink"; no chapter has the procedure. Add a sourced procedure to §2.12 or drop the promise.
- Pins: `requires-python = ">=3.14,<3.15"` in ch01/ch02 vs `">=3.14"` elsewhere (pick one); ch23 README installs shellcheck via apt (unpinned) — say so or use the container; 07-tech-stack.md lists pins no exercise uses (hypothesis, psycopg, redis-py, client_golang, go-oidc, gobreaker, six actions) — mark them "mentioned only".
- Indirect golang.org/x/text versions differ across modules (v0.29–v0.42): harmless, note in appendix C (already listed there).
- Hookline: x/sync v0.22.0 → v0.23.0 (asked of the capstone writer in M6).
- Hookline M5 smoke step in §38.5.4 uses `docker compose exec postgres psql ... UPDATE`, not run as written (writer ran the docker exec equivalent): run it once.
