package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"strings"

	"github.com/samyak-jain/agora_backend/graph/generated"
	"github.com/samyak-jain/agora_backend/graph/model"
	"github.com/samyak-jain/agora_backend/middleware"
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
)

func (r *mutationResolver) CreateChannel(ctx context.Context, title string, enablePstn *bool) (*model.ShareResponse, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	var pstnResponse *model.Pstn
	var newChannel *models.Channel

	hostPhrase, err := utils.GenerateUUID()
	if err != nil {
		return nil, err
	}

	viewPhrase, err := utils.GenerateUUID()
	if err != nil {
		return nil, err
	}

	channelName, err := utils.GenerateUUID()
	if err != nil {
		return nil, err
	}

	channel := strings.ReplaceAll(channelName, "-", "")

	if *enablePstn {
		dtmfResult, err := utils.GenerateDTMF()
		if err != nil {
			return nil, err
		}

		pstnResponse = &model.Pstn{
			Number: "+17018052515",
			Dtmf:   *dtmfResult,
		}

		newChannel = &models.Channel{
			Title:            title,
			Name:             channel,
			HostPassphrase:   hostPhrase,
			ViewerPassphrase: viewPhrase,
			DTMF:             *dtmfResult,
			Hosts:            *authUser,
		}

	} else {
		pstnResponse = nil
		newChannel = &models.Channel{
			Title:            title,
			Name:             channel,
			HostPassphrase:   hostPhrase,
			ViewerPassphrase: viewPhrase,
			Hosts:            *authUser,
		}
	}

	r.DB.Create(newChannel)
	r.DB.Save(newChannel)

	return &model.ShareResponse{
		Passphrase: &model.Passphrase{
			Host: &hostPhrase,
			View: viewPhrase,
		},
		Title:   title,
		Channel: channel,
		Pstn:    pstnResponse,
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

func (r *mutationResolver) StartRecordingSession(ctx context.Context, passphrase string) (string, error) {
	var channelData models.Channel
	var host bool

	if passphrase == "" {
		return "", errors.New("Passphrase cannot be empty")
	}

	if r.DB.Where("host_passphrase = ?", passphrase).First(&channelData).RecordNotFound() {
		if r.DB.Where("viewer_passphrase = ?", passphrase).First(&channelData).RecordNotFound() {
			return "", errors.New("Invalid passphrase")
		}

		host = false
	} else {
		host = true
	}

	if !host {
		return "", errors.New("Unauthorised to record channel")
	}

	recorder := &utils.Recorder{}
	recorder.Channel = channelData.Name

	err := recorder.Acquire()
	if err != nil {
		return "", err
	}

	err = recorder.Start()
	if err != nil {
		return "", err
	}

	if err := r.DB.Model(&channelData).Update("Recording", &models.Recording{
		UID: recorder.UID,
		SID: recorder.SID,
		RID: recorder.RID,
	}).Error; err != nil {
		return "", err
	}

	return "success", nil
}

func (r *mutationResolver) StopRecordingSession(ctx context.Context, passphrase string) (string, error) {
	var channelData models.Channel
	var host bool

	if passphrase == "" {
		return "", errors.New("Passphrase cannot be empty")
	}

	if r.DB.Where("host_passphrase = ?", passphrase).First(&channelData).RecordNotFound() {
		if r.DB.Where("viewer_passphrase = ?", passphrase).First(&channelData).RecordNotFound() {
			return "", errors.New("Invalid passphrase")
		}

		host = false
	} else {
		host = true
	}

	if !host {
		return "", errors.New("Unauthorised to record channel")
	}

	var record models.Recording
	r.DB.Model(&channelData).Related(&record)

	err := utils.Stop(channelData.Name, record.UID, record.RID, record.SID)
	if err != nil {
		return "", err
	}

	return "success", nil
}

func (r *mutationResolver) LogoutSession(ctx context.Context, token string) ([]string, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	tokenIndex := -1

	r.DB.Preload("Tokens").Find(&authUser)

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

func (r *queryResolver) JoinChannel(ctx context.Context, passphrase string) (*model.Session, error) {
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

	mainUser, err := utils.GenerateUserCredentials(channelData.Name, true)
	if err != nil {
		return nil, err
	}

	screenShare, err := utils.GenerateUserCredentials(channelData.Name, false)
	if err != nil {
		return nil, err
	}

	return &model.Session{
		Title:       channelData.Title,
		Channel:     channelData.Name,
		IsHost:      host,
		MainUser:    mainUser,
		ScreenShare: screenShare,
	}, nil
}

func (r *queryResolver) Share(ctx context.Context, passphrase string) (*model.ShareResponse, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

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

	var hostPassphrase *string
	if host {
		hostPassphrase = &channelData.HostPassphrase
	} else {
		hostPassphrase = nil
	}

	var pstnResult *model.Pstn
	if channelData.DTMF != "" {
		pstnResult = &model.Pstn{
			Number: "+17018052515",
			Dtmf:   channelData.DTMF,
		}
	} else {
		pstnResult = nil
	}

	return &model.ShareResponse{
		Passphrase: &model.Passphrase{
			Host: hostPassphrase,
			View: channelData.ViewerPassphrase,
		},
		Channel: channelData.Name,
		Title:   channelData.Title,
		Pstn:    pstnResult,
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

	r.DB.Preload("Tokens").Find(&authUser)

	return authUser.GetAllTokens(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
