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

	models "github.com/OWASP/Amass/v3/models"
	"github.com/caffix/netmap"
	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

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

func (p *Postgres) InsertExecution(domains []string) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("Error connecting to database: %v\n", err)
	}

	execution := models.Execution{
		Domains: strings.Join(domains, ", "),
	}

	result := db.Create(&execution)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating execution: %v\n", result.Error)
	}

	return execution.ID, nil
}

func (p *Postgres) insertExecutionLog(db *gorm.DB, assetId int64, execId int64) (int64, error) {
	exec_log := models.ExecutionLog{
		ExecutionID: execId,
		AssetID:     assetId,
	}

	log_result := db.Create(&exec_log)
	if log_result.Error != nil {
		return assetId, fmt.Errorf("Error creating execution log: %v\n", log_result.Error)
	}

	return assetId, nil
}

// Create FQDN if it does not exist, otherwise return the ID of the existing FQDN
func (p *Postgres) InsertFQDN(info InsertInfo, fqdn FQDN) (int64, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return 0, fmt.Errorf("Could not get sql connection: %s", err)
	}

	var asset models.Asset
	err = db.Where("content->>'name' = ?", fqdn.Name).First(&asset).Error
	if err == nil {
		return asset.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("Failed to check if FDQN already exists: %v\n", err)
	}

	tld, _ := publicsuffix.PublicSuffix(fqdn.Name)
	domain, err := publicsuffix.EffectiveTLDPlusOne(fqdn.Name)
	if err != nil {
		return 0, fmt.Errorf("InsertFQDN: Failed to obtain a valid domain name for %s", fqdn.Name)
	}

	in_fqdn := models.FQDN{
		Name: domain,
		Tld:  tld}

	fqdn_content, err := json.Marshal(in_fqdn)
	if err != nil {
		return 0, fmt.Errorf("Error marshalling FQDN: %v\n", err)
	}

	in_asset := models.Asset{
		Type:    "fqdn",
		Content: datatypes.JSON(fqdn_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating FQDN asset: %v\n", result.Error)
	}

	return p.insertExecutionLog(db, in_asset.ID, info.EventID)
}

// Create relation between two assets
func (p *Postgres) insertRelation(from_id int64, to_id int64, relation string) error {
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
func (p *Postgres) insertFQDNRelation(info InsertInfo, fqdn string, target string, relation string) error {
	target_fqdn := FQDN{target}
	target_id, err := p.InsertFQDN(info, target_fqdn)
	if err != nil {
		return fmt.Errorf("InsertFQDN: Failed to insert target FQDN %s: %v", target, err)
	}

	in_fqdn := FQDN{fqdn}
	fqdn_id, err := p.InsertFQDN(info, in_fqdn)
	if err != nil {
		return fmt.Errorf("InsertFQDN: Failed to insert FQDN %s: %v", fqdn, err)
	}

	err = p.insertRelation(fqdn_id, target_id, relation)

	return nil
}

func (p *Postgres) InsertCNAME(info InsertInfo, dns DNSRecord) error {
	return p.insertFQDNRelation(info, dns.Fqdn, dns.Target, "cname_record")
}

func (p *Postgres) InsertPTR(info InsertInfo, dns DNSRecord) error {
	return p.insertFQDNRelation(info, dns.Fqdn, dns.Target, "ptr_record")
}

func (p *Postgres) InsertSRV(info InsertInfo, srv Service) error {
	err := p.insertFQDNRelation(info, srv.Service, srv.Fqdn, "service")
	if err != nil {
		return err
	}

	return p.insertFQDNRelation(info, srv.Service, srv.Target, "srv_record")
}

func (p *Postgres) InsertNS(info InsertInfo, dns DNSRecord) error {
	return p.insertFQDNRelation(info, dns.Fqdn, dns.Target, "ns_record")
}

func (p *Postgres) InsertMX(info InsertInfo, dns DNSRecord) error {
	return p.insertFQDNRelation(info, dns.Fqdn, dns.Target, "mx_record")
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
func (p *Postgres) insertIPAddr(info InsertInfo, addr string, version string) (int64, error) {
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
		Type:    "ip",
		Content: datatypes.JSON(ip_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating IP asset: %v\n", result.Error)
	}

	return p.insertExecutionLog(db, in_asset.ID, info.EventID)
}

// Create Netblock asset
func (p *Postgres) insertNetblock(info InsertInfo, cidr string, version string) (int64, error) {
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
		Type:    "netblock",
		Content: datatypes.JSON(netblock_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating netblock asset: %v\n", result.Error)
	}

	return p.insertExecutionLog(db, in_asset.ID, info.EventID)
}

// Create Autonomous System asset
func (p *Postgres) insertAS(info InsertInfo, asn int64) (int64, error) {
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
		Type:    "as",
		Content: datatypes.JSON(as_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating AS asset: %v\n", result.Error)
	}

	return p.insertExecutionLog(db, in_asset.ID, info.EventID)
}

// Create RIR Organization asset
func (p *Postgres) insertRIROrg(info InsertInfo, rir string) (int64, error) {
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
		Type:    "rirorg",
		Content: datatypes.JSON(riro_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		return 0, fmt.Errorf("Error creating RIROrg asset: %v\n", result.Error)
	}

	return p.insertExecutionLog(db, in_asset.ID, info.EventID)
}

func (p *Postgres) InsertInfrastructure(info InsertInfo, infra Infrastructure) error {
	ip_id, err := p.insertIPAddr(info, infra.Address, "")
	if err != nil {
		return fmt.Errorf("Error inserting IP address: %v\n", err)
	}

	netblock_id, err := p.insertNetblock(info, infra.Cidr, "")
	if err != nil {
		return fmt.Errorf("Error inserting netblock: %v\n", err)
	}

	as_id, err := p.insertAS(info, int64(infra.Asn))
	if err != nil {
		return fmt.Errorf("Error inserting AS: %v\n", err)
	}

	rirorg_id, err := p.insertRIROrg(info, infra.Description)
	if err != nil {
		return fmt.Errorf("Error inserting RIR Org: %v\n", err)
	}

	err = p.insertRelation(netblock_id, ip_id, "contains")
	if err != nil {
		return fmt.Errorf("Error inserting relation: %v\n", err)
	}

	err = p.insertRelation(as_id, netblock_id, "announces")
	if err != nil {
		return fmt.Errorf("Error inserting relation: %v\n", err)
	}

	err = p.insertRelation(as_id, rirorg_id, "managed_by")
	if err != nil {
		return fmt.Errorf("Error inserting relation: %v\n", err)
	}

	return nil
}

func (p *Postgres) InsertA(info InsertInfo, record HostRecord) error {
	fqdn := FQDN{record.Fqdn}
	fqdn_id, err := p.InsertFQDN(info, fqdn)
	if err != nil {
		return fmt.Errorf("Error inserting FQDN: %v\n", err)
	}

	ip_id, err := p.insertIPAddr(info, record.Address, "v4")
	if err != nil {
		return fmt.Errorf("Error inserting IP address: %v\n", err)
	}

	err = p.insertRelation(fqdn_id, ip_id, "a_record")

	return nil
}

func (p *Postgres) InsertAAAA(info InsertInfo, record HostRecord) error {
	fqdn := FQDN{record.Fqdn}
	fqdn_id, err := p.InsertFQDN(info, fqdn)
	if err != nil {
		return fmt.Errorf("Error inserting FQDN: %v\n", err)
	}

	ip_id, err := p.insertIPAddr(info, record.Address, "v6")
	if err != nil {
		return fmt.Errorf("Error inserting IP address: %v\n", err)
	}

	err = p.insertRelation(fqdn_id, ip_id, "aaaa_record")

	return nil
}

func (p *Postgres) IsCNAMENode(ctx context.Context, fqdn string) (bool, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return false, fmt.Errorf("Could not get sql connection: %s", err)
	}

	var count int64
	err = db.
		Model(&models.Asset{}).
		Where("content->>'name' = ?", fqdn).
		Joins("JOIN relations ON relations.from_asset_id = assets.id").
		Where("relations.type = ?", "cname").
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("Could not get count of CNAME relations: %s", err)
	}
	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}

}

func (p *Postgres) Migrate(ctx context.Context, graph *netmap.Graph) error {
	return nil
}

func (p *Postgres) EventFQDNs(ctx context.Context, execID int64) []string {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil
	}

	var fqdnNames []string
	err = db.
		Model(&models.Asset{}).
		Select("content->>'name'").
		Joins("JOIN execution_logs ON execution_logs.asset_id = assets.id").
		Where("execution_logs.execution_id = ? AND assets.type = ?", execID, "fqdn").
		Scan(&fqdnNames).Error

	if err != nil {
		return nil
	}

	return fqdnNames
}

func (p *Postgres) NamesToAddrs(ctx context.Context, execID int64, names ...string) ([]*NameAddrPair, error) {
	db, err := gorm.Open(postgres.Open(p.db.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("Could not get sql connection: %s", err)
	}

	var results []*NameAddrPair
	err = db.
		Model(&models.Asset{}).
		Select("assets.content->>'name' as name, assets_r.content->>'address' as addr").
		Joins("JOIN relations ON relations.from_asset_id = assets.id").
		Joins("JOIN assets as assets_r ON assets_r.id = relations.to_asset_id").
		Where("assets.type = ? AND assets_r.type = ? AND (relations.type = ? OR relations.type = ?)", "fqdn", "ip", "a_record", "aaaa_record").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("Could not get name-address relations: %s", err)
	}

	return results, nil
}
