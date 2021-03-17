package migrations

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/samyak-jain/agora_backend/utils"
	"log"
)

// RunMigration runs the schema migrations
func RunMigration() {
	utils.SetupConfig()
	//db, err := models.CreateDB(utils.GetDBURL())
	//if err != nil {
	//	log.Print(err)
	//	return
	//}
	//
	//defer db.Close()

	//db.AutoMigrate(&models.User{}, &models.Channel{}, &models.Token{})
	m, err := migrate.New(
		"file://migrations/",
		utils.GetDBURL())
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
