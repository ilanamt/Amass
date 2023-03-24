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
	Primary         bool   `ini:"primary"`
	System          string `ini:"system"`
	URL             string `ini:"url"`
	Host            string `ini:"host"`
	Port            string `ini:"port"`
	Username        string `ini:"username"`
	Password        string `ini:"password"`
	DBName          string `ini:"database"`
	SSLMode         string `ini:"sslmode"`
	MigrationsPath  string `ini:"migrations_path"`
	MigrationsTable string `ini:"migrations_table"`
	Options         string `ini:"options"`
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

type Store interface {
	getAppliedMigrationsCount() (int, error)
	getMigrationsSource() *migrate.FileMigrationSource
	getSqlConnection() (*sql.DB, error)
	CreateDatabaseIfNotExists() error
	DropDatabase() error
	GetPendingMigrationsCount() (int, error)
	IsDatabaseCreated() (bool, error)
	RunInitMigration() error
	RunMigrations() error
	InsertFQDN(InsertInfo, FQDN) (int64, error)
	InsertCNAME(InsertInfo, DNSRecord) error
	InsertPTR(InsertInfo, DNSRecord) error
	InsertSRV(InsertInfo, Service) error
	InsertNS(InsertInfo, DNSRecord) error
	InsertMX(InsertInfo, DNSRecord) error
	InsertInfrastructure(InsertInfo, Infrastructure) error
	InsertA(InsertInfo, HostRecord) error
	InsertAAAA(InsertInfo, HostRecord) error
	IsCNAMENode(context.Context, string) (bool, error)
	InsertExecution([]string) (int64, error)
	Migrate(context.Context, *netmap.Graph) error
	EventFQDNs(context.Context, int64) []string
	NamesToAddrs(context.Context, int64, ...string) ([]*NameAddrPair, error)
}

type InsertInfo struct {
	Ctx     context.Context
	Source  string
	EventID int64
}

type FQDN struct {
	Name string
}

type DNSRecord struct {
	Fqdn   string
	Target string
}

type HostRecord struct {
	Fqdn    string
	Address string
}

type Service struct {
	Fqdn    string
	Target  string
	Service string
}

type Infrastructure struct {
	Asn         int
	Description string
	Address     string
	Cidr        string
}

type NameAddrPair struct {
	Name string
	Addr string
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
		if mgr == nil || mgr.(*Cayley).db != db {
			mgr = NewCayley(db)
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
