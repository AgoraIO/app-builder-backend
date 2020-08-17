package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/samyak-jain/agora_backend/graph/generated"
	"github.com/samyak-jain/agora_backend/graph/model"
	"github.com/samyak-jain/agora_backend/middleware"
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
	uuid "github.com/satori/go.uuid"
)

func (r *mutationResolver) CreateChannel(ctx context.Context, channel string, password *model.PasswordInput, enableLink *bool) (*model.ShareResponse, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	var channelData models.Channel
	if !r.DB.Where("name = ?", channel).First(&channelData).RecordNotFound() {
		return nil, errors.New("Channel name already taken")
	}

	var hostPhrase string
	var viewPhrase string

	usephrase := false
	usepass := false

	if *enableLink {
		usephrase = true
		hostPhrase = uuid.NewV4().String()
		viewPhrase = uuid.NewV4().String()
	}

	if password != nil {
		usepass = true
	} else {
		password = &model.PasswordInput{Host: "", View: ""}
	}

	newChannel := &models.Channel{
		Name:             channel,
		UsePassword:      usepass,
		HostPassword:     password.Host,
		ViewerPassword:   password.View,
		UsePassphrase:    usephrase,
		HostPassphrase:   hostPhrase,
		ViewerPassphrase: viewPhrase,
		Creator:          *authUser,
	}

	r.DB.Create(newChannel)
	// r.DB.Model(newChannel).Association("Creator").Append(authUser)

	passwordResponse := model.Password(*password)

	return &model.ShareResponse{
		Password: &passwordResponse,
		Passphrase: &model.Passphrase{
			Host: hostPhrase,
			View: viewPhrase,
		},
	}, nil
}

func (r *mutationResolver) UpdateUserName(ctx context.Context, name string) (*model.User, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	user := &models.User{Email: authUser.Email}
	if err := r.DB.Model(&user).Update("name", name).Error; err != nil {
		return nil, err
	}

	return &model.User{
		Name:  name,
		Email: authUser.Email,
	}, nil
}

func (r *mutationResolver) LogoutSession(ctx context.Context, token string) ([]string, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	tokenIndex := -1

	for index := range authUser.Tokens {
		if authUser.Tokens[index].TokenID == token {
			tokenIndex = index
		}
	}

	if tokenIndex == -1 {
		return nil, errors.New("Invalid token")
	}

	r.DB.Delete(&models.Token{
		TokenID: token,
	})

	return authUser.GetAllTokens(), nil
}

func (r *mutationResolver) LogoutAllSessions(ctx context.Context) (*string, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	r.DB.Where("user_email = ?", authUser.Email).Delete(models.Token{})

	return nil, nil
}

func (r *queryResolver) JoinChannel(ctx context.Context, channel string, password string) (*model.Session, error) {
	var channelData models.Channel
	if err := r.DB.Where("name = ?", channel).First(&channelData).Error; err != nil {
		return nil, err
	}

	if !channelData.UsePassword {
		return nil, errors.New("Cannot join using password")
	}

	uid := int(rand.Uint32())
	rtcToken, err := utils.GetRtcToken(channel, uid)
	if err != nil {
		return nil, err
	}

	rtmToken, err := utils.GetRtmToken(fmt.Sprint(uid))
	if err != nil {
		return nil, err
	}

	var host bool
	if password == channelData.HostPassword {
		host = true
	} else if password == channelData.ViewerPassword {
		host = false
	} else {
		return nil, errors.New("Invalid Password")
	}

	return &model.Session{
		Channel: &channel,
		Rtc:     rtcToken,
		Rtm:     rtmToken,
		UID:     uid,
		IsHost:  host,
	}, nil
}

func (r *queryResolver) JoinChannelWithPassphrase(ctx context.Context, passphrase string) (*model.Session, error) {
	var channelData models.Channel
	var host bool

	if passphrase == "" {
		return nil, errors.New("Passphrase cannot be empty")
	}

	if r.DB.Where("host_passphrase = ?", passphrase).First(&channelData).RecordNotFound() {
		if r.DB.Where("viewer_passphrase = ?", passphrase).First(&channelData).RecordNotFound() {
			return nil, errors.New("Invalid passphrase")
		}

		host = false
	} else {
		host = true
	}

	if !channelData.UsePassphrase {
		return nil, errors.New("Cannot login using passphrase")
	}

	uid := int(rand.Uint32())
	rtcToken, err := utils.GetRtcToken(channelData.Name, uid)
	if err != nil {
		return nil, err
	}

	rtmToken, err := utils.GetRtmToken(fmt.Sprint(uid))
	if err != nil {
		return nil, err
	}

	return &model.Session{
		Channel: &channelData.Name,
		Rtc:     rtcToken,
		Rtm:     rtmToken,
		UID:     uid,
		IsHost:  host,
	}, nil
}

func (r *queryResolver) Share(ctx context.Context, channel string) (*model.ShareResponse, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	var channelData models.Channel
	var userData models.User

	if err := r.DB.Where("name = ?", channel).First(&channelData).Related(&userData).Error; err != nil {
		return nil, err
	}

	if userData.Email != authUser.Email {
		return nil, errors.New("Unauthorized Access")
	}

	return &model.ShareResponse{
		Password:   &model.Password{Host: channelData.HostPassword, View: channelData.ViewerPassword},
		Passphrase: &model.Passphrase{Host: channelData.HostPassphrase, View: channelData.ViewerPassphrase},
	}, nil
}

func (r *queryResolver) GetUser(ctx context.Context) (*model.User, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	return &model.User{
		Name:  authUser.Name,
		Email: authUser.Email,
	}, nil
}

func (r *queryResolver) GetSessions(ctx context.Context) ([]string, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	return authUser.GetAllTokens(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
