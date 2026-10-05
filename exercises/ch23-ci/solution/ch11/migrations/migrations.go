// Package migrations holds the lab schema as goose SQL migrations, embedded in the binary
// so that tests and services migrate without depending on the working directory.
package migrations

import "embed"

// FS contains every NNNNN_name.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
