package service

import (
	"context"
	"jkt48lab/model"
	"jkt48lab/repository"
)

type IDNLiveService interface {
	FindAll(ctx context.Context) ([]model.Live, error)
	Find(ctx context.Context, username string) (model.Live, error)
	SendNotification(ctx context.Context, onLives *model.OnLives, live model.Live)
}

type IDNLiveServiceImpl struct {
	IDNLiveRepository repository.IDNLiveRepository
	DiscordService    DiscordService
}

func (service *IDNLiveServiceImpl) SendNotification(ctx context.Context, onLives *model.OnLives, live model.Live) {
	isNotified, isStreaming := service.IDNLiveRepository.IsNotified(ctx, onLives, live.MemberUsername)
	if isNotified == false && isStreaming == true {
		service.DiscordService.SendStartNotification(live)
	}
	if isNotified == true && isStreaming == false {
		service.DiscordService.SendEndNotification(live, 0)
	}
}

func (service *IDNLiveServiceImpl) FindAll(ctx context.Context) ([]model.Live, error) {
	lives, err := service.IDNLiveRepository.FindAll(ctx)
	return lives, err
}

func (service *IDNLiveServiceImpl) Find(ctx context.Context, username string) (model.Live, error) {
	live, err := service.IDNLiveRepository.Find(ctx, username)
	return live, err
}

func NewIDNLiveService(liveRepository repository.IDNLiveRepository) IDNLiveService {
	return &IDNLiveServiceImpl{
		IDNLiveRepository: liveRepository,
		DiscordService:    NewDiscordService(),
	}
}
