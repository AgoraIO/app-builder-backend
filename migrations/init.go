package migrations

import (
	"github.com/samyak-jain/agora_backend/models"
)

// RunMigration runs the schema migrations
func RunMigration(db *models.Database) {
	db.AutoMigrate(&models.User{}, &models.Channel{}, &models.Token{})
}
