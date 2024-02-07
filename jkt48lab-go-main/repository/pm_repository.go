package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"io"
	"jkt48lab/helper"
	"jkt48lab/model"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type PMRepository interface {
	FindAllMessages(ctx context.Context, accessToken string) ([]model.PMMessage, error)
	FindBirthdayMessage(ctx context.Context, accessToken string, username string) (model.PMBirthdayMessage, error)
	FindLastMessage(ctx context.Context) model.PMMessage
	UpdateLastMessage(ctx context.Context, message model.PMMessage)
	FindRankings(ctx context.Context, accessToken string, from string, until string, max int) []model.PMRanking
	FindMessagesByUserIdByDate(ctx context.Context, accessToken string, userId string, from string, until string) []model.PMMessageByUserId
}

type PMRepositoryImpl struct {
}

func NewPMRepository() PMRepository {
	return &PMRepositoryImpl{}
}

func (repository *PMRepositoryImpl) FindAllMessages(ctx context.Context, accessToken string) ([]model.PMMessage, error) {
	query, _ := json.Marshal(map[string]any{
		"query": `
			query MessagesByUpdateAt {
				messagesByUpdateAt(type: "message", sortDirection: DESC, limit: 2000) {
					items {
						id
						message
						channelId
						createdAt
						updatedAt
						userMessagesId
						type
						author {
							givenName
							familyName
							nickname
							profileImage
						}
					}
				}
			}
    	`,
	})
	resp, err := helper.GraphQLRequest("https://xzqpphzvbzhzvpke6ojjzvbpjq.appsync-api.ap-southeast-1.amazonaws.com/graphql", query, accessToken)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var pmMessageResponses model.PMMessageResponses
	json.Unmarshal(body, &pmMessageResponses)
	messages := pmMessageResponses.Data.MessagesByUpdateAt.Items

	if len(messages) == 0 {
		return messages, errors.New("Invalid Authorization Token")
	}

	return messages, nil
}

var BirthDayMessages map[string]model.PMBirthdayMessage

func (repository *PMRepositoryImpl) FindBirthdayMessage(ctx context.Context, accessToken string, username string) (model.PMBirthdayMessage, error) {
	val, ok := BirthDayMessages[username]
	if ok {
		return val, nil
	}
	BirthDayMessages = make(map[string]model.PMBirthdayMessage)
	query, _ := json.Marshal(map[string]any{
		"query": `
			query ListBirthdayMessageTemplates {
				listBirthdayMessageTemplates {
					items {
						id
						message
						author {
							familyName
							givenName
							nickname
						}
					}
				}
			}

    	`,
	})
	resp, err := helper.GraphQLRequest("https://xzqpphzvbzhzvpke6ojjzvbpjq.appsync-api.ap-southeast-1.amazonaws.com/graphql", query, accessToken)
	if err != nil {
		log.Println(err)
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
	}

	var pmBirthdayResponses model.PMBirthdayResponses
	var birthdayMessage model.PMBirthdayMessage
	if err := json.Unmarshal(body, &pmBirthdayResponses); err != nil {
		log.Println("Gagal mengubah JSON ke PMResponses")
	}

	messages := pmBirthdayResponses.Data.ListBirthdayMessageTemplates.Items
	if len(messages) == 0 {
		log.Println("[ERROR] Birthday Messages EMPTY")
		return birthdayMessage, errors.New("Invalid Authorization Token")
	}

	for _, message := range messages {
		if message.Author.Nickname == username {
			birthdayMessage = message
		}
	}

	BirthDayMessages[birthdayMessage.Author.Nickname] = birthdayMessage
	return birthdayMessage, nil
}

func (repository *PMRepositoryImpl) FindAllBirthdayMessages(ctx context.Context, accessToken string) (map[string]model.PMBirthdayMessage, error) {
	BirthDayMessages = make(map[string]model.PMBirthdayMessage)
	query, _ := json.Marshal(map[string]any{
		"query": `
			query ListBirthdayMessageTemplates {
				listBirthdayMessageTemplates {
					items {
						id
						message
						author {
							id
							familyName
							givenName
							nickname
							profileImage
						}
					}
				}
			}

    	`,
	})
	resp, err := helper.GraphQLRequest("https://xzqpphzvbzhzvpke6ojjzvbpjq.appsync-api.ap-southeast-1.amazonaws.com/graphql", query, accessToken)
	if err != nil {
		log.Println(err)
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
	}
	var pmBirthdayResponses model.PMBirthdayResponses
	json.Unmarshal(body, &pmBirthdayResponses)
	messages := pmBirthdayResponses.Data.ListBirthdayMessageTemplates.Items
	for _, message := range messages {
		BirthDayMessages[message.Author.Nickname] = message
	}
	return BirthDayMessages, nil
}

func (repository *PMRepositoryImpl) FindLastMessage(ctx context.Context) model.PMMessage {
	var file []byte
	for {
		f, err := os.ReadFile("./data/pm_last_message.json")
		if err == nil {
			file = f
			break
		} else {
			log.Println("[Error] Failed read pm_last_message.json. Retrying...")
		}
	}
	var lastPmMessages model.PMMessage
	for {
		err := json.Unmarshal(file, &lastPmMessages)
		if err == nil {
			break
		} else {
			log.Println("[Error] Failed unmarshal pm_last_message.json. Retrying...")
		}
	}
	return lastPmMessages
}

func (repository *PMRepositoryImpl) UpdateLastMessage(ctx context.Context, message model.PMMessage) {
	marshal, _ := json.MarshalIndent(message, "", "  ")
	os.WriteFile("./data/pm_last_message.json", marshal, os.ModePerm)
}

func (repository *PMRepositoryImpl) FindRankings(ctx context.Context, accessToken string, from string, until string, max int) []model.PMRanking {
	log.Println(fmt.Sprintf("[INFO] Preparing Update Ranking (%s - %s). Please wait...", from, until))
	rankings := make(map[string]model.PMRanking)
	birthdayMessages, _ := repository.FindAllBirthdayMessages(ctx, accessToken)
	for _, birthdayMessage := range birthdayMessages {
		messagesByDate := repository.FindMessagesByUserIdByDate(ctx, accessToken, birthdayMessage.Author.Id, from, until)
		for _, message := range messagesByDate {
			if message.Message == birthdayMessage.Message {
				continue
			}
			similarityCheck := strutil.Similarity(birthdayMessage.Message, message.Message, metrics.NewJaccard())
			if similarityCheck < 0.89 {
				var points int
				var textCount int
				var voiceCount int
				var imageCount int
				if strings.Contains(message.Message, "ucarecdn.com") {
					resp, err := http.Get(message.Message)
					if err != nil {
						log.Println(err)
					}
					contentType := resp.Header.Get("Content-Type")
					if contentType == "audio/x-m4a" {
						voiceCount++
						points += 3
					} else {
						imageCount++
						points += 2
					}
				} else {
					textCount++
					points += 1
				}
				rankings[birthdayMessage.Author.Nickname] = model.PMRanking{
					Name:         fmt.Sprintf("%s (%s %s)", birthdayMessage.Author.Nickname, birthdayMessage.Author.GivenName, birthdayMessage.Author.FamilyName),
					Count:        rankings[birthdayMessage.Author.Nickname].Count + 1,
					TextCount:    rankings[birthdayMessage.Author.Nickname].TextCount + textCount,
					VoiceCount:   rankings[birthdayMessage.Author.Nickname].VoiceCount + voiceCount,
					ImageCount:   rankings[birthdayMessage.Author.Nickname].ImageCount + imageCount,
					Points:       rankings[birthdayMessage.Author.Nickname].Points + points,
					ProfileImage: birthdayMessage.Author.ProfileImage,
				}
			} else {
				continue
			}
		}
	}
	keys := make([]string, 0, len(rankings))
	for key := range rankings {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return rankings[keys[i]].Points > rankings[keys[j]].Points
	})

	var sortedRanks []model.PMRanking
	sortedRanks = helper.SortMapToSlicePMStats(rankings, keys, max)

	log.Println("Success", len(sortedRanks))
	return sortedRanks
}

func (repository *PMRepositoryImpl) FindMessagesByUserIdByDate(ctx context.Context, accessToken string, userId string, from string, until string) []model.PMMessageByUserId {
	query, _ := json.Marshal(map[string]any{
		"query": fmt.Sprintf(`
			query GetUser {
				getUser(id: "%s") {
					messages(
						filter: { createdAt: { between: ["%s", "%s"] } }
						limit: 100000000
					) {
						items {
							message
						}
					}
				}
			}
    	`, userId, from, until),
	})
	resp, err := helper.GraphQLRequest("https://xzqpphzvbzhzvpke6ojjzvbpjq.appsync-api.ap-southeast-1.amazonaws.com/graphql", query, accessToken)
	if err != nil {
		log.Println(err)
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var pMMessageByUserIdResponses model.PMMessageByUserIdResponses
	json.Unmarshal(body, &pMMessageByUserIdResponses)
	messages := pMMessageByUserIdResponses.Data.GetUser.Messages.Items

	return messages
}
