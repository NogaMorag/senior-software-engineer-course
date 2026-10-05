// Package migrations holds the users schema and the four steps that rename
// users.fullname to users.display_name, as goose SQL migrations embedded in the binary.
package migrations

import "embed"

// FS contains every NNNNN_name.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
