package embed

import "embed"

//go:embed db/migrations/postgres/*
var PostgresMigrations embed.FS

func GetPostgresMigrationSource() embed.FS {
	return PostgresMigrations
}
