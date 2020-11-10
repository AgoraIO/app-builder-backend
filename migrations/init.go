package migrations

import (
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
)

// RunMigration runs the schema migrations
func RunMigration(db *models.Database) {
	utils.SetupConfig()
	db.AutoMigrate(&models.User{}, &models.Channel{}, &models.Token{})
}
