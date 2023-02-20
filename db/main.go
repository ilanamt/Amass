package main

import (
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// main is a placeholder while additional commands are being developed
// TODO: this should be removed when the db cli command supports init, migrate, and
// other related commands
func main() {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations/postgres",
	}

	dsn := "host=localhost dbname=asset_db user=myuser password=mypass sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		fmt.Printf("Error opening db\n")
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("Error creating generic database\n")
	}

	n, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Printf("Error making migrations\n")
	}
	fmt.Printf("Applied %d migrations!\n", n)
}
