package graph

//go:generate go run github.com/99designs/gqlgen

import (
	"github.com/samyak-jain/agora_backend/pkg/models"
	"github.com/samyak-jain/agora_backend/utils"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is used for state management
type Resolver struct {
	DB     *models.Database
	Logger *utils.Logger
}
