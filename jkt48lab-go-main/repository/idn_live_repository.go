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
	"mvdan.cc/xurls/v2"
	"strings"
	"time"
)

type IDNLiveRepository interface {
	FindAll(ctx context.Context) ([]model.Live, error)
	Find(ctx context.Context, username string) (model.Live, error)
	IsNotified(ctx context.Context, onLives *model.OnLives, username string) (bool, bool)
}

type IDNLiveRepositoryImpl struct {
}

func (repository *IDNLiveRepositoryImpl) FindAll(ctx context.Context) ([]model.Live, error) {
	var lives []model.Live
	var page = 1
	for {
		resp, _ := helper.GraphQLIDN(page)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}
		resp.Body.Close()

		var result model.LiveIDNResponses
		if err := json.Unmarshal(body, &result); err != nil {
			log.Println("Gagal mengubah JSON ke LiveIDNResponses")
		}

		if len(result.Data.GetLiveStream) > 0 {
			for _, data := range result.Data.GetLiveStream {
				if !strings.Contains(data.Creator.Username, "jkt48") {
					continue
				}
				if data.Status == "scheduled" {
					continue
				}

				xurl := xurls.Relaxed()

				resp, _ := helper.Fetch(data.PlaybackUrl)
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
				}
				resp.Body.Close()
				playbackUrl := xurl.FindAllString(string(body), -1)

				if len(playbackUrl) < 2 {
					break
				}

				startedAt, _ := time.Parse("2024-01-09T06:27:22+07:00", data.LiveAt)
				live := model.Live{
					MemberUsername:    data.Creator.Username,
					MemberDisplayName: data.Creator.Name,
					Platform:          "IDN",
					Title:             data.Title,
					StreamUrl:         playbackUrl[1],
					Views:             data.ViewCount,
					ImageUrl:          data.ImageUrl,
					StartedAt:         int(startedAt.Unix()),
				}
				lives = append(lives, live)

			}
			page++
		} else {
			break
		}
	}
	return lives, nil
}

func (repository *IDNLiveRepositoryImpl) Find(ctx context.Context, username string) (model.Live, error) {
	var live model.Live
	lives, _ := repository.FindAll(ctx)
	for _, l := range lives {
		if l.MemberUsername == username {
			live = l
			return live, nil
		}
	}
	return live, errors.New(fmt.Sprintf("%s sedang tidak live"))
}

func (repository *IDNLiveRepositoryImpl) IsNotified(ctx context.Context, onLives *model.OnLives, username string) (bool, bool) {
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

func NewIDNLiveRepository() IDNLiveRepository {
	return &IDNLiveRepositoryImpl{}
}
