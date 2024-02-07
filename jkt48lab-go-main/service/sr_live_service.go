package service

import (
	"context"
	"jkt48lab/model"
	"jkt48lab/repository"
)

type SRLiveService interface {
	FindAll(ctx context.Context) ([]model.Live, error)
	Find(ctx context.Context, username string) (model.Live, error)
	SendNotification(ctx context.Context, onLives *model.OnLives, live model.Live)
	CountGiftPoints(ctx context.Context, live model.Live) int
}

type SRLiveServiceImpl struct {
	SRLiveRepository repository.SRLiveRepository
	DiscordService   DiscordService
}

func (service *SRLiveServiceImpl) SendNotification(ctx context.Context, onLives *model.OnLives, live model.Live) {
	isNotified, isStreaming := service.SRLiveRepository.IsNotified(ctx, onLives, live.MemberUsername)
	if isNotified == false && isStreaming == true {
		service.DiscordService.SendStartNotification(live)
	}
	if isNotified == true && isStreaming == false {
		points := service.CountGiftPoints(ctx, live)
		service.DiscordService.SendEndNotification(live, points)
	}
}

func (service *SRLiveServiceImpl) FindAll(ctx context.Context) ([]model.Live, error) {
	lives, err := service.SRLiveRepository.FindAll(ctx)
	return lives, err
}

func (service *SRLiveServiceImpl) Find(ctx context.Context, username string) (model.Live, error) {
	live, err := service.SRLiveRepository.Find(ctx, username)
	return live, err
}

func (service *SRLiveServiceImpl) CountGiftPoints(ctx context.Context, live model.Live) int {
	result := service.SRLiveRepository.FindAllGift(ctx, live)
	var total int
	for _, data := range result {
		total += data.Point * data.Num
	}
	return total
}

func NewSRLiveService(liveRepository repository.SRLiveRepository, discordService DiscordService) SRLiveService {
	return &SRLiveServiceImpl{
		SRLiveRepository: liveRepository,
		DiscordService:   discordService,
	}
}
