// Package migrations embeds the goose SQL migration files for the db package.
package migrations

import "embed"

// FS holds the embedded goose migration files.
//
//go:embed *.sql
var FS embed.FS
