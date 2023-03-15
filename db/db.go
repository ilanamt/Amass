// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/caffix/netmap"
	migrate "github.com/rubenv/sql-migrate"
)

// Database contains values required for managing databases.
type Database struct {
	Primary         bool          `ini:"primary"`
	System          string        `ini:"system"`
	URL             string        `ini:"url"`
	Host            string        `ini:"host"`
	Port            string        `ini:"port"`
	Graph           *netmap.Graph `ini:"graph"`
	Username        string        `ini:"username"`
	Password        string        `ini:"password"`
	DBName          string        `ini:"database"`
	SSLMode         string        `ini:"sslmode"`
	MigrationsPath  string        `ini:"migrations_path"`
	MigrationsTable string        `ini:"migrations_table"`
	Options         string        `ini:"options"`
}

func (d Database) String() string {
	s := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s", d.Host, d.Port, d.DBName, d.SSLMode)
	if d.Username != "" {
		s += fmt.Sprintf(" user=%s", d.Username)
	}
	if d.Password != "" {
		s += fmt.Sprintf(" password=%s", d.Password)
	}
	return s
}

type SQLStore interface {
	getAppliedMigrationsCount() (int, error)
	getMigrationsSource() *migrate.FileMigrationSource
	getSqlConnection() (*sql.DB, error)
	CreateDatabaseIfNotExists() error
	DropDatabase() error
	GetPendingMigrationsCount() (int, error)
	IsDatabaseCreated() (bool, error)
	RunInitMigration() error
	RunMigrations() error
}

type Store interface {
	UpsertFQDN(context.Context, string, string, int64) (int64, error)
	UpsertCNAME(context.Context, string, string, string, int64) error
	UpsertPTR(context.Context, string, string, string, int64) error
	UpsertSRV(context.Context, string, string, string, string, int64) error
	UpsertNS(context.Context, string, string, string, int64) error
	UpsertMX(context.Context, string, string, string, int64) error
	UpsertInfrastructure(context.Context, int, string, string, string, string, int64) error
	UpsertA(context.Context, string, string, string, int64) error
	UpsertAAAA(context.Context, string, string, string, int64) error
}

func GetDatabaseManager(db *Database) Store {
	var mgr Store

	switch db.System {
	case "postgres":
		if mgr == nil || mgr.(*Postgres).db != db {
			mgr = &Postgres{db: db}
		}
		return mgr
	case "cayley":
		if mgr == nil || mgr.(*Cayley).db != db || mgr.(*Cayley).graph != db.Graph {
			mgr = &Cayley{db: db, graph: db.Graph}
		}
		return mgr
	default:
		// Temporary Default
		// TODO: Update to local store
		if mgr == nil || mgr.(*Postgres).db != db {
			mgr = &Postgres{db: db}
		}
		return mgr
	}

	return nil
}
