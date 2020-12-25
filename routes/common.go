package routes

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/samyak-jain/agora_backend/utils"

	"github.com/markbates/goth/gothic"
	"github.com/samyak-jain/agora_backend/models"
)

// Router refers to all the oauth endpoints
type Router struct {
	DB *models.Database
}

// TokenTemplate is a struct that will be used to template the token into the html that will be served for Desktop and Mobile
type TokenTemplate struct {
	Token  string
	Scheme string
}

// Handler is the handler that will do most of the heavy lifting for OAuth
func Handler(w http.ResponseWriter, r *http.Request, db *models.Database, platform string) (*string, *string, error) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Msg("Could not parse form request")
		return nil, nil, err
	}

	state := r.FormValue("state")
	fmt.Println("state : ", state)
	decodedState, err := url.QueryUnescape(state)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Msg("Could not url decode state")
		return nil, nil, err
	}
	fmt.Println("decodedState : ", decodedState)

	parsedState, err := url.ParseQuery(decodedState)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("Could not parse deocoded state")
		return nil, nil, err
	}

	redirect := parsedState.Get("redirect")

	fmt.Println("redirect :", redirect)

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Msg("Couldn't complete request")
		return nil, nil, err
	}
	fmt.Println("\nRawData :", user)
	fmt.Println("\n\nF Name :", user.FirstName)
	fmt.Println("L Name :", user.LastName)
	fmt.Println("Location :", user.Location)
	fmt.Println("UserID :", user.UserID)
	fmt.Println("Email :", user.Email)

	bearerToken, err := utils.GenerateUUID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("Could not generate bearer token")
		return nil, nil, err
	}

	var userData models.User
	if db.Where("email = ?", user.Email).First(&userData).RecordNotFound() {
		db.Create(&models.User{
			Name:  user.Name,
			Email: user.Email,
			Tokens: []models.Token{{
				TokenID: bearerToken,
			}},
		})
	} else {
		db.Model(&userData).Association("Tokens").Append(models.Token{
			TokenID: bearerToken,
		})
	}
	return &redirect, &bearerToken, nil
}
