// Package migrations holds Hookline's schema as goose SQL migrations (chapter 11, section 11.6), embedded
// in the binaries so that hookline-admin and the tests migrate without depending on the working directory.
//
// The four files create the whole data model of chapter 37, section 37.4, at milestone M2. Deliveries and
// attempts stay unused until milestone M5, but creating them now means no later milestone rewrites a table.
package migrations

import "embed"

// FS contains every NNNNN_name.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
