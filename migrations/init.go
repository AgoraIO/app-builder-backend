// ********************************************
// Copyright © 2021 Agora Lab, Inc., all rights reserved.
// AppBuilder and all associated components, source code, APIs, services, and documentation
// (the “Materials”) are owned by Agora Lab, Inc. and its licensors.  The Materials may not be
// accessed, used, modified, or distributed for any purpose without a license from Agora Lab, Inc.
// Use without a license or in violation of any license terms and conditions (including use for
// any purpose competitive to Agora Lab, Inc.’s business) is strictly prohibited.  For more
// information visit https://appbuilder.agora.io.
// *********************************************

package migrations

import (
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/spf13/viper"
)

// RunMigration runs the schema migrations
func RunMigration(config *string) {
	if config == nil {
		configDir := flag.String("config", "..", "Directory which contains the config.json")
		utils.SetupConfig(configDir)
	}

	println("Running Migrations...")
	m, err := migrate.New(
		"file://migrations/migrations",
		viper.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Print(err)
	}
	// Uncomment the below lines for migrate down!
	//if err := m.Down(); err != nil {
	//	log.Fatal(err)
	//}
}
