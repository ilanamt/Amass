// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/OWASP/Amass/v3/config"
	amassdb "github.com/OWASP/Amass/v3/db"
	"github.com/fatih/color"
)

func runDBInitSubCommand(cfg *config.Config) {
	database, err := openSQLDatabase(cfg)
	if err != nil {
		r.Fprintf(color.Error, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	manager := amassdb.GetDatabaseManager(database).(amassdb.Store)

	if err := manager.CreateDatabaseIfNotExists(); err != nil {
		r.Fprintf(color.Error, "Failed to create database: %v\n", err)
		os.Exit(1)
	}
	if err := manager.RunInitMigration(); err != nil {
		r.Fprintf(color.Error, "Failed to initialize migrations: %v\n", err)
		os.Exit(1)
	}
	// Check if there are pending migrations
	if plannedMigrations, err := manager.GetPendingMigrationsCount(); err != nil {
		r.Fprintf(color.Error, "Failed to get pending migrations count: %v\n", err)
		os.Exit(1)
	} else if plannedMigrations == 1 {
		fgY.Fprintf(color.Output, "WARNING: There is a pending migration that has not been applied!\n%s\n", dbUpgradeMsg)
	} else if plannedMigrations > 1 {
		fgY.Fprintf(color.Output, "WARNING: There are %d pending migrations that have not been applied!\n%s\n", plannedMigrations, dbUpgradeMsg)
	}
}
