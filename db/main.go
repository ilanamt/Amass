package main

import (
	"fmt"

	"github.com/OWASP/Amass/v3/db/models"
	"github.com/jackc/pgx/pgtype"
	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// main is a placeholder while additional commands are being developed
// TODO: this should be removed when the db cli command supports init, migrate, and
// other related commands
func main() {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations/postgres",
	}

	dsn := "host=localhost dbname=asset_db user=myuser password=mypass sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("Error opening db\n")
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("Error creating generic database\n")
	}

	p, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Down)
	if err != nil {
		fmt.Printf("Error making migrations down\n")
		fmt.Printf("Error is %v\n", err)
	}
	fmt.Printf("Applied %d migrations!\n", p)

	n, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Printf("Error making migrations\n")
	}
	fmt.Printf("Applied %d migrations!\n\n", n)

	fmt.Printf("Creating enum\n")
	enum := models.EnumExecution{}
	result := db.Create(&enum)
	if result.Error != nil {
		fmt.Printf("Error creating enum: %v\n", result.Error)
	} else {
		fmt.Printf("Enum ID is %d\n", enum.ID)
	}

	fmt.Printf("Creating FQDN asset\n")
	asset1 := models.Asset{
		EnumExecutionID: enum.ID,
		Type:            "FQDN",
		Content: pgtype.JSONB{
			Bytes: []byte(`{"fqdn": "example.com"}`)}}
	result1 := db.Create(&asset1)
	if result1.Error != nil {
		fmt.Printf("Error creating asset: %v\n", result1.Error)
	} else {
		fmt.Printf("Asset1 ID is %d\n", asset1.ID)
	}

	fmt.Printf("Creating ASN asset\n")
	asset2 := models.Asset{
		EnumExecutionID: enum.ID,
		Type:            "ASN",
		Content: pgtype.JSONB{
			Bytes: []byte(`{"asn": 1234}`)}}
	result2 := db.Create(&asset2)
	if result2.Error != nil {
		fmt.Printf("Error creating asset: %v\n", result2.Error)
	} else {
		fmt.Printf("Asset2 ID is %d\n", asset2.ID)
	}

	fmt.Printf("Creating relation\n")
	relation := models.Relation{
		Type:        "related",
		FromAssetID: asset1.ID,
		ToAssetID:   asset2.ID}
	result3 := db.Create(&relation)
	if result3.Error != nil {
		fmt.Printf("Error creating relation: %v\n", result3.Error)
	} else {
		fmt.Printf("Enum ID is %d\n", relation.ID)
	}
}
