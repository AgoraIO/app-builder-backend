package main

import (
	"log"

	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
)

func main() {
	utils.SetupConfig()
	db, err := models.CreateDB(utils.GetDBURL())
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	db.AutoMigrate(&models.User{}, &models.Channel{})
}
