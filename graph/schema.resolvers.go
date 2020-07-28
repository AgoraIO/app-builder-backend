package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	stdtime "time"

	"github.com/google/uuid"
	"github.com/samyak-jain/agora_backend/graph/generated"
	"github.com/samyak-jain/agora_backend/graph/model"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/samyak-jain/agora_backend/utils/rtctoken"
	"github.com/samyak-jain/agora_backend/utils/rtmtoken"
)

var config utils.AgoraConfig = utils.GetConfig()

func (r *mutationResolver) Login(ctx context.Context, token string) (string, error) {
	bin, err := uuid.New().MarshalBinary()
	if err != nil {
		return "", err
	}

	cookie := base64.StdEncoding.EncodeToString(bin)
	return cookie, nil
}

func (r *queryResolver) Rtc(ctx context.Context, channel string, uid int, role *model.AgoraRole, time *int) (string, error) {
	var RtcRole rtctoken.Role = rtctoken.RolePublisher

	if role != nil && *role == model.AgoraRoleSubscriber {
		RtcRole = rtctoken.RoleSubscriber
	}

	var expireTimestamp uint32 = 0

	if time != nil {
		currentTimestamp := uint32(stdtime.Now().UTC().Unix())
		expireTimestamp = currentTimestamp + uint32(*time)
	}

	return rtctoken.BuildTokenWithUID(config.AppID, config.AppCertificate, channel, uint32(uid), RtcRole, expireTimestamp)
}

func (r *queryResolver) Rtm(ctx context.Context, user string, time *int) (string, error) {
	var expireTimestamp uint32 = 0

	if time != nil {
		currentTimestamp := uint32(stdtime.Now().UTC().Unix())
		expireTimestamp = currentTimestamp + uint32(*time)
	}

	return rtmtoken.BuildToken(config.AppID, config.AppCertificate, user, rtmtoken.RoleRtmUser, expireTimestamp)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
