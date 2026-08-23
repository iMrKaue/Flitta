package migrations

import "embed"

// Files contém as migrations SQL empacotadas no binário.
//
//go:embed *.sql
var Files embed.FS
