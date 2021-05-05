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
		configDir := flag.String("config", ".", "Directory which contains the config.json")
		utils.SetupConfig(configDir)
	}

	m, err := migrate.New(
		"file://migrations/",
		viper.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
	// Uncomment the below lines for migrate down!
	//if err := m.Down(); err != nil {
	//	log.Fatal(err)
	//}
}
