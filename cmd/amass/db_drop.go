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

func runDBDropCommand(cfg *config.Config) {
	database, err := openSQLDatabase(cfg)
	if err != nil {
		r.Fprintf(color.Error, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	manager := amassdb.GetDatabaseManager(database).(amassdb.Store)

	if err := manager.DropDatabase(); err != nil {
		r.Fprintf(color.Error, "Failed to drop database: %v\n", err)
		os.Exit(1)
	}
}
