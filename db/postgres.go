// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/OWASP/Amass/v3/db/models"
	models "github.com/OWASP/Amass/v3/models"
	"golang.org/x/net/publicsuffix"
)

// Postgres implements the Store and SQLStore interfaces
type Postgres struct {
	db *Database
}

// getAppliedMigrationsAmount returns the number of already applied migrations.
// If the function is unable to interact w/ the database,
// 0 is returned with the error that occured in interacting with the database.
func (p *Postgres) getAppliedMigrationsCount() (int, error) {
	if p.db.MigrationsTable != "" {
		migrate.SetTable(p.db.MigrationsTable)
	}

	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return 0, fmt.Errorf("could not get applied migrations amount: %s", err)
	}

	appliedMigrations, err := migrate.GetMigrationRecords(sqlDb, "postgres")
	if err != nil {
		return 0, fmt.Errorf("could not get applied migrations amount: %s", err)
	}

	return len(appliedMigrations), nil
}

// getSqlConnection returns a generic SQL database interface using the GORM interface.
func (p *Postgres) getSqlConnection() (*sql.DB, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("could not get sql connection: %s", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("could not get sql connection: %s", err)
	}

	return sqlDb, nil
}

func (p *Postgres) getMigrationsSource() *migrate.FileMigrationSource {
	if p.db.MigrationsTable != "" {
		migrate.SetTable(p.db.MigrationsTable)
	}
	return &migrate.FileMigrationSource{Dir: p.db.MigrationsPath}
}

// getPendingMigrationsCount returns the number of pending migrations to be applied.
// If the function is unable to interact w/ the database,
// 0 is returned with the error that occured in interacting with the database.
func (p *Postgres) GetPendingMigrationsCount() (int, error) {
	migrationsSource := p.getMigrationsSource()
	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return 0, fmt.Errorf("could not get pending migrations count: %s", err)
	}

	plannedMigrations, _, err := migrate.PlanMigration(sqlDb, "postgres", migrationsSource, migrate.Up, math.MaxInt32)
	if err != nil {
		return 0, fmt.Errorf("could not get pending migrations count: %s", err)
	}

	return len(plannedMigrations), nil
}

// CreateDatabaseIfNotExists will attempt to create the configured database if it does not exist.
// If the connection to the database fails it will return an error.
func (p *Postgres) CreateDatabaseIfNotExists() error {
	doesExist, err := p.IsDatabaseCreated()
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %s", err)
	}
	if doesExist {
		return nil
	}
	fmt.Println("Database does not exist. Creating database...")
	// Try to create database connecting to default postgres database on same host with same user
	db_copy := *p.db
	db_copy.DBName = "postgres"
	pg_db, err := gorm.Open(postgres.Open(db_copy.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("could not connect to postgres database: %s", err)
	}
	stmt := fmt.Sprintf("CREATE DATABASE %s;", p.db.DBName)
	if rs := pg_db.Exec(stmt); rs.Error != nil {
		return fmt.Errorf("could not create database: %s", rs.Error)
	}
	return nil
}

// DropDatabase is a destructive command that will drop configured database.
// If the connection to the default Postgres database fails, it will return an error.
func (p *Postgres) DropDatabase() error {
	// Try to create database connecting to default postgres database on same host with same user
	db_copy := *p.db
	db_copy.DBName = "postgres"
	pg_db, err := gorm.Open(postgres.Open(db_copy.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("could not connect to postgres database: %s", err)
	}
	stmt := fmt.Sprintf("DROP DATABASE %s;", p.db.DBName)
	if rs := pg_db.Exec(stmt); rs.Error != nil {
		return fmt.Errorf("could not drop database: %s", rs.Error)
	}
	return nil
}

// IsDatabaseCreated checks that the provided Postgres database already exists.
func (p *Postgres) IsDatabaseCreated() (bool, error) {
	_, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("database \"%s\" does not exist", p.db.DBName)) {
			return false, nil
		} else {
			return false, fmt.Errorf("could not connect to database: %s", err)
		}
	}
	return true, nil
}

// RunInitMigration runs migrations if no migrations have been applied before.
func (p *Postgres) RunInitMigration() error {
	appliedMigrationsAmount, err := p.getAppliedMigrationsCount()
	if err != nil {
		return fmt.Errorf("could not run init migration: %s", err)
	}
	if appliedMigrationsAmount >= 1 {
		fmt.Println("Database already initialized.")
		return nil
	}

	migrations := p.getMigrationsSource()
	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return fmt.Errorf("could not run init migration: %s", err)
	}

	maxMigrations := 0 // 0 means no limit
	n, err := migrate.ExecMax(sqlDb, "postgres", migrations, migrate.Up, maxMigrations)
	if err != nil {
		return fmt.Errorf("could not execute %d migrations: %s", n, err)
	}

	if n == 1 {
		fmt.Printf("Applied %d migration!\n", n)
	} else {
		fmt.Printf("Applied %d migrations!\n", n)
	}
	return nil
}

// RunMigrations runs all pending migrations.
func (p *Postgres) RunMigrations() error {
	migrations := p.getMigrationsSource()

	sqlDb, err := p.getSqlConnection()
	if err != nil {
		return fmt.Errorf("could not run migrations: %s", err)
	}

	n, err := migrate.Exec(sqlDb, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("could not run migrations: %s", err)
	}

	if n == 0 {
		fmt.Println("No migrations to apply.")
	} else if n == 1 {
		fmt.Printf("Applied %d migration!\n", n)
	} else {
		fmt.Printf("Applied %d migrations!\n", n)
	}

	return nil
}

// Create FQDN if it does not exist, otherwise return the ID of the existing FQDN
func (p *Postgres) UpsertFQDN(ctx context.Context, name string, source string, eventID int64) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("Could not get sql connection: %s", err)
	}

	var asset models.Asset
	err = db.Where("content->>'name' = ?", "example.com").First(&asset).Error
	if err == nil {
		return asset.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("Failed to check if FDQN already exists: %v\n", err)
	}

	tld, _ := publicsuffix.PublicSuffix(name)
	domain, err := publicsuffix.EffectiveTLDPlusOne(name)
	if err != nil {
		return 0, fmt.Errorf("UpsertFQDN: Failed to obtain a valid domain name for %s", name)
	}

	fqdn := models.FQDN{
		Name: domain,
		Tld:  tld}

	fqdn_content, err := json.Marshal(fqdn)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling FQDN: %v\n", err)
	}

	in_asset := models.Asset{
		EnumExecutionID: eventID,
		Type:            "fqdn",
		Content:         datatypes.JSON(fqdn_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating FQDN asset: %v\n", result.Error)
	}

	return in_asset.ID, nil
}

// Create relation between two assets
func (p *Postgres) upsertRelation(from_id int64, to_id int64, relation string) error {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return fmt.Errorf("could not get sql connection: %s", err)
	}

	in_relation := models.Relation{
		FromAssetID: from_id,
		ToAssetID:   to_id,
		Type:        relation,
	}

	relation_result := db.Create(&in_relation)
	if relation_result.Error != nil {
		return fmt.Errorf("Error creating relation: %v\n", relation_result.Error)
	}

	return nil
}

// Create relation between two FQDN assets
func (p *Postgres) upsertFQDNRelation(ctx context.Context, fqdn string, target string, source string, eventID int64, relation string) error {
	target_id, err := p.UpsertFQDN(ctx, target, source, eventID)
	if err != nil {
		return fmt.Errorf("UpsertFQDN: Failed to upsert target FQDN %s: %v", target, err)
	}

	fqdn_id, err := p.UpsertFQDN(ctx, fqdn, source, eventID)
	if err != nil {
		return fmt.Errorf("UpsertFQDN: Failed to upsert FQDN %s: %v", fqdn, err)
	}

	err = p.upsertRelation(fqdn_id, target_id, relation)

	return nil
}

func (p *Postgres) UpsertCNAME(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return p.upsertFQDNRelation(ctx, fqdn, target, source, eventID, "cname_record")
}

func (p *Postgres) UpsertPTR(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return p.upsertFQDNRelation(ctx, fqdn, target, source, eventID, "ptr_record")
}

func (p *Postgres) UpsertSRV(ctx context.Context, fqdn string, service string, target string, source string, eventID int64) error {
	err := p.upsertFQDNRelation(ctx, service, fqdn, source, eventID, "service")
	if err != nil {
		return err
	}

	return p.upsertFQDNRelation(ctx, service, target, source, eventID, "srv_record")
}

func (p *Postgres) UpsertNS(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return p.upsertFQDNRelation(ctx, fqdn, target, source, eventID, "ns_record")
}

func (p *Postgres) UpsertMX(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return p.upsertFQDNRelation(ctx, fqdn, target, source, eventID, "mx_record")
}

func determineIpVersion(ip net.IP) string {
	if ip.To4() != nil {
		return "v4"
	} else if ip.To16() != nil {
		return "v6"
	}
	return ""
}

// Create IP Address asset
func (p *Postgres) upsertIPAddr(ctx context.Context, addr string, source string, eventID int64, version string) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("could not get sql connection: %s", err)
	}

	ip_addr := net.ParseIP(addr)
	if version == "" {
		version = determineIpVersion(ip_addr)
	}

	ip := models.IPAddress{
		Address: ip_addr,
		Type:    version}

	ip_content, err := json.Marshal(ip)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling IP: %v\n", err)
	}

	in_asset := models.Asset{
		EnumExecutionID: eventID,
		Type:            "ip",
		Content:         datatypes.JSON(ip_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating IP asset: %v\n", result.Error)
	}

	return in_asset.ID, nil
}

// Create Netblock asset
func (p *Postgres) upsertNetblock(ctx context.Context, cidr string, source string, eventID int64, version string) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("could not get sql connection: %s", err)
	}

	ip, cidr_val, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, fmt.Errorf("Error parsing CIDR: %v\n", err)
	}

	if version == "" {
		version = determineIpVersion(ip)
	}

	netblock := models.Netblock{
		Cidr: *cidr_val,
		Type: version,
	}

	netblock_content, err := json.Marshal(netblock)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling netblock: %v\n", err)
	}

	in_asset := models.Asset{
		EnumExecutionID: eventID,
		Type:            "netblock",
		Content:         datatypes.JSON(netblock_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating netblock asset: %v\n", result.Error)
	}

	return in_asset.ID, nil
}

// Create Autonomous System asset
func (p *Postgres) upsertAS(ctx context.Context, asn int64, source string, eventID int64) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("could not get sql connection: %s", err)
	}

	as := models.AutonomousSystem{
		Number: asn}

	as_content, err := json.Marshal(as)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling AS: %v\n", err)
	}

	in_asset := models.Asset{
		EnumExecutionID: eventID,
		Type:            "as",
		Content:         datatypes.JSON(as_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating AS asset: %v\n", result.Error)
	}

	return in_asset.ID, nil
}

// Create RIR Organization asset
func (p *Postgres) upsertRIROrg(ctx context.Context, rir string, source string, eventID int64) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("could not get sql connection: %s", err)
	}

	riro := models.RIROrganization{
		Name:  rir,
		RIRId: "unknown",
		RIR:   "unknown",
	}

	riro_content, err := json.Marshal(riro)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling RIR: %v\n", err)
	}

	in_asset := models.Asset{
		EnumExecutionID: eventID,
		Type:            "rirorg",
		Content:         datatypes.JSON(riro_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating RIROrg asset: %v\n", result.Error)
	}

	return in_asset.ID, nil
}

func (p *Postgres) UpsertInfrastructure(ctx context.Context, asn int, desc string, addr string, cidr string, source string, eventID int64) error {
	ip_id, err := p.upsertIPAddr(ctx, addr, source, eventID, "")
	if err != nil {
		return fmt.Errorf("Error upserting IP address: %v\n", err)
	}

	netblock_id, err := p.upsertNetblock(ctx, cidr, source, eventID, "")
	if err != nil {
		return fmt.Errorf("Error upserting netblock: %v\n", err)
	}

	as_id, err := p.upsertAS(ctx, int64(asn), source, eventID)
	if err != nil {
		return fmt.Errorf("Error upserting AS: %v\n", err)
	}

	rirorg_id, err := p.upsertRIROrg(ctx, desc, source, eventID)
	if err != nil {
		return fmt.Errorf("Error upserting RIR Org: %v\n", err)
	}

	err = p.upsertRelation(netblock_id, ip_id, "contains")
	if err != nil {
		return fmt.Errorf("Error upserting relation: %v\n", err)
	}

	err = p.upsertRelation(as_id, netblock_id, "announces")
	if err != nil {
		return fmt.Errorf("Error upserting relation: %v\n", err)
	}

	err = p.upsertRelation(as_id, rirorg_id, "managed_by")
	if err != nil {
		return fmt.Errorf("Error upserting relation: %v\n", err)
	}

	return nil
}

func (p *Postgres) UpsertA(ctx context.Context, fqdn string, addr string, source string, eventID int64) error {
	fqdn_id, err := p.UpsertFQDN(ctx, fqdn, source, eventID)
	if err != nil {
		return fmt.Errorf("Error upserting FQDN: %v\n", err)
	}

	ip_id, err := p.upsertIPAddr(ctx, addr, source, eventID, "v4")
	if err != nil {
		return fmt.Errorf("Error upserting IP address: %v\n", err)
	}

	err = p.upsertRelation(fqdn_id, ip_id, "a_record")

	return nil
}

func (p *Postgres) UpsertAAAA(ctx context.Context, fqdn string, addr string, source string, eventID int64) error {
	fqdn_id, err := p.UpsertFQDN(ctx, fqdn, source, eventID)
	if err != nil {
		return fmt.Errorf("Error upserting FQDN: %v\n", err)
	}

	ip_id, err := p.upsertIPAddr(ctx, addr, source, eventID, "v6")
	if err != nil {
		return fmt.Errorf("Error upserting IP address: %v\n", err)
	}

	err = p.upsertRelation(fqdn_id, ip_id, "aaaa_record")

	return nil
}
