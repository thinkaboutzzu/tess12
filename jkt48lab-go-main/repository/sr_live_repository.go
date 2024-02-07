package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"jkt48lab/helper"
	"jkt48lab/model"
	"log"
	"strings"
)

type SRLiveRepository interface {
	FindAll(ctx context.Context) ([]model.Live, error)
	Find(ctx context.Context, username string) (model.Live, error)
	IsNotified(ctx context.Context, onLives *model.OnLives, username string) (bool, bool)
	FindAllGift(ctx context.Context, live model.Live) []model.LiveShowroomGift
}

type SRLiveRepositoryImpl struct {
}

func (repository *SRLiveRepositoryImpl) FindAll(ctx context.Context) ([]model.Live, error) {
	resp, err := helper.Fetch("https://www.showroom-live.com/api/live/onlives")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var result model.LiveShowroomResponses
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("Gagal mengubah JSON ke LiveShowroomResponses")
	}

	var lives []model.Live
	if len(result.OnLives) > 0 {
		for _, data := range result.OnLives[0].Lives {
			if data.PremiumRoomType == 1 {
				continue
			}
			if !strings.Contains(data.RoomUrlKey, "JKT48") {
				continue
			}
			if data.RoomId == 0 {
				continue
			}
			resp, err := helper.Fetch(fmt.Sprintf("https://www.showroom-live.com/api/live/streaming_url?abr_available=1&room_id=%d", data.RoomId))
			if err != nil {
				log.Println(err)
			}
			body, err := io.ReadAll(resp.Body)

			var result model.LiveShowroomStreamingUrlResponses
			if err := json.Unmarshal(body, &result); err != nil {
				log.Println("Gagal mengubah JSON ke LiveShowroomStreamingUrlResponses")
			}
			resp.Body.Close()

			if len(result.StreamingUrlList) > 0 {
				live := model.Live{
					MemberUsername:    data.RoomUrlKey,
					MemberDisplayName: data.MainName,
					Platform:          "Showroom",
					Title:             fmt.Sprintf("%s Live", data.MainName),
					StreamUrl:         result.StreamingUrlList[1].Url,
					Views:             data.ViewNum,
					ImageUrl:          data.Image,
					StartedAt:         data.StartedAt,
					RoomId:            data.RoomId,
				}
				lives = append(lives, live)
			}

		}
	}
	return lives, nil
}

func (repository *SRLiveRepositoryImpl) Find(ctx context.Context, username string) (model.Live, error) {
	var live model.Live
	lives, _ := repository.FindAll(ctx)
	for _, l := range lives {
		if l.MemberUsername == username {
			live = l
			return live, nil
		}
	}
	return live, errors.New(fmt.Sprintf("%s sedang tidak live", username))
}

func (repository *SRLiveRepositoryImpl) IsNotified(ctx context.Context, onLives *model.OnLives, username string) (bool, bool) {
	// return isNotified, isStreaming
	_, err := repository.Find(ctx, username)
	if err != nil {
		// Member sedang tidak live
		if helper.Contains(onLives.MemberOnLives, username) {
			onLives.MemberOnLives = helper.RemoveStringFromSlice(onLives.MemberOnLives, username)
			return true, false
		}
	}
	if !helper.Contains(onLives.MemberOnLives, username) {
		onLives.MemberOnLives = append(onLives.MemberOnLives, username)
		return false, true
	} else {
		return true, true
	}
}

func (repository *SRLiveRepositoryImpl) FindAllGift(ctx context.Context, live model.Live) []model.LiveShowroomGift {
	resp, err := helper.Fetch(fmt.Sprintf("https://www.showroom-live.com/api/live/gift_list?room_id=%d", live.RoomId))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var giftList model.LiveShowroomGiftListResponses
	if err := json.Unmarshal(body, &giftList); err != nil {
		log.Println("Gagal mengubah JSON ke LiveShowroomGiftListResponses")
	}

	resp, err = helper.Fetch(fmt.Sprintf("https://www.showroom-live.com/api/live/gift_log?room_id=%d", live.RoomId))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)

	var giftLog model.LiveShowroomGiftLogResponses
	if err := json.Unmarshal(body, &giftLog); err != nil {
		log.Println("Gagal mengubah JSON ke LiveShowroomGiftLogResponses")
	}

	log.Println(len(giftList.Normal) + len(giftList.Enquete))
	log.Println(len(giftLog.GiftLog))

	var gifts []model.LiveShowroomGift
	for _, gift := range giftLog.GiftLog {
		for _, ge := range giftList.Enquete {
			if gift.GiftId == ge.GiftId {
				gifts = append(gifts, model.LiveShowroomGift{
					GiftType: ge.GiftType,
					Image:    ge.Image,
					Free:     ge.Free,
					Point:    ge.Point,
					GiftName: ge.GiftName,
					Num:      gift.Num,
					UserId:   gift.UserId,
				})
			}
		}
		for _, gn := range giftList.Normal {
			if gift.GiftId == gn.GiftId {
				gifts = append(gifts, model.LiveShowroomGift{
					GiftType: gn.GiftType,
					Image:    gn.Image,
					Free:     gn.Free,
					Point:    gn.Point,
					GiftName: gn.GiftName,
					Num:      gift.Num,
					UserId:   gift.UserId,
				})
			}
		}
	}
	return gifts
}

func NewSRLiveRepository() SRLiveRepository {
	return &SRLiveRepositoryImpl{}
}
