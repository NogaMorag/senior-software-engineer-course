// Package migrations holds Hookline's schema as goose SQL migrations (chapter 11, section 11.6), embedded
// in the binaries so that hookline-admin and the tests migrate without depending on the working directory.
//
// The first four files create the whole data model of chapter 37, section 37.4, at milestone M2.
// Milestone M5 adds one column (00005, attempt_base), the additive change chapter 29, section 29.5
// designs, and milestone M6 one table (00006, leader_terms) for the relay's leader election. No
// milestone rewrites a table.
package migrations

import "embed"

// FS contains every NNNNN_name.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
