package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"io"
	"jkt48lab/helper"
	"jkt48lab/model"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type DiscordService interface {
	GetSession() *discordgo.Session
	GetWebhooks() []Webhook
	SendStartNotification(live model.Live)
	SendEndNotification(live model.Live, points int)
	SendPrivateMessages(message []model.PMMessage)
	GetPrivateMessageChannelID(username string) string
	GetPrivateMessageRoleID(username string) string
	SendRankUpdate(ranks []model.PMRanking, updateType string, byteImage []byte)
}

type DiscordServiceImpl struct {
}

type Webhook struct {
	Id    string `json:"id"`
	Token string `json:"token"`
}

func NewDiscordService() DiscordService {
	return &DiscordServiceImpl{}
}

var BOT *discordgo.Session

func (service *DiscordServiceImpl) GetSession() *discordgo.Session {
	if BOT == nil {
		// Not Created
		godotenv.Load()
		token := os.Getenv("DISCORD_TOKEN")
		bot, err := discordgo.New("Bot " + token)
		if err != nil {
			log.Fatal(err)
		}
		defer bot.Close()
		BOT = bot
	}
	return BOT
}

func (service *DiscordServiceImpl) GetWebhooks() []Webhook {
	jsonFile, err := os.Open("./data/notification_webhooks.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	webhooksJson, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var webhooks []Webhook
	err = json.Unmarshal(webhooksJson, &webhooks)
	if err != nil {
		log.Fatal(err)
	}

	return webhooks
}

func (service *DiscordServiceImpl) SendStartNotification(live model.Live) {
	bot := service.GetSession()
	webhooks := service.GetWebhooks()

	var liveUrl string
	if live.Platform == "IDN" {
		liveUrl = fmt.Sprintf("https://jkt48.safatanc.com/live/idn/%s", live.MemberUsername)
	} else if live.Platform == "Showroom" {
		liveUrl = fmt.Sprintf("https://jkt48.safatanc.com/live/sr/%s", live.MemberUsername)
	}

	var liveUrlOriginal string
	if live.Platform == "IDN" {
		liveUrlOriginal = fmt.Sprintf("https://idn.app/%s", live.MemberUsername)
	} else if live.Platform == "Showroom" {
		liveUrlOriginal = fmt.Sprintf("https://showroom-live.com/r/%s", live.MemberUsername)
	}

	for _, webhook := range webhooks {
		embed := &discordgo.MessageEmbed{
			URL:         "https://jkt48.safatanc.com",
			Title:       "Notifikasi Live",
			Description: fmt.Sprintf("**%s** sedang live di `%s`", live.MemberDisplayName, live.Platform),
			Color:       0xccec1c,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "JKT48Lab by safatanc.com",
			},
			Image: &discordgo.MessageEmbedImage{
				URL: live.ImageUrl,
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Tonton di JKT48Lab",
					Value:  fmt.Sprintf("[Klik disini](%s)", liveUrl),
					Inline: true,
				},
				{
					Name:   fmt.Sprintf("Tonton di %s", live.Platform),
					Value:  fmt.Sprintf("[Klik disini](%s)", liveUrlOriginal),
					Inline: true,
				},
				{
					Name:   " ",
					Value:  "[**Join discord JKT48Lab**](https://discord_service.gg/dVgmJfmXc2)",
					Inline: false,
				},
			},
		}

		_, err := bot.WebhookExecute(webhook.Id, webhook.Token, false, &discordgo.WebhookParams{
			Username: "JKT48Lab",
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (service *DiscordServiceImpl) SendEndNotification(live model.Live, points int) {
	bot := service.GetSession()
	webhooks := service.GetWebhooks()

	for _, webhook := range webhooks {
		embed := &discordgo.MessageEmbed{
			URL:         "https://jkt48.safatanc.com",
			Title:       "Notifikasi Live Berakhir",
			Description: fmt.Sprintf("Live **%s** di `%s` telah berakhir", live.MemberDisplayName, live.Platform),
			Color:       0xccec1c,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "JKT48Lab by safatanc.com",
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Total Gift",
					Value:  fmt.Sprintf("🎁 %d points\n💰 ± Rp%d", points, points*105),
					Inline: true,
				},
				{
					Name:   " ",
					Value:  "[**Join discord JKT48Lab**](https://discord_service.gg/dVgmJfmXc2)",
					Inline: false,
				},
			},
		}
		_, err := bot.WebhookExecute(webhook.Id, webhook.Token, false, &discordgo.WebhookParams{
			Username: "JKT48Lab",
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (service *DiscordServiceImpl) SendPrivateMessages(messages []model.PMMessage) {
	bot := service.GetSession()
	pmQueue := make(map[string][]*discordgo.MessageEmbed)
	for _, message := range messages {
		imageUrl := ""
		content := message.Message
		if strings.Contains(message.Message, "ucarecdn.com") {
			resp, err := http.Get(message.Message)
			if err != nil {
				log.Println(err)
			}
			if resp.Header.Get("Content-Type") == "audio/x-m4a" {
				content = fmt.Sprintf("[**Voice Note** (klik untuk membuka)](%s)", message.Message)
			} else {
				imageUrl = message.Message
				content = ""
			}
		}
		resp, err := http.Get(message.Author.ProfileImage)
		if err != nil {
			log.Println(err)
		}
		if resp.StatusCode == 404 {
			message.Author.ProfileImage = "https://jkt48.primesse.me/_next/image?url=%2Fimages%2Ficon.png&w=1080&q=75"
		}
		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    fmt.Sprintf("PM %s (%s %s)", message.Author.Nickname, message.Author.GivenName, message.Author.FamilyName),
				IconURL: message.Author.ProfileImage,
			},
			Description: content,
			Color:       0x2C2F33,
			Image: &discordgo.MessageEmbedImage{
				URL: imageUrl,
			},
			Timestamp: message.CreatedAt,
		}
		pmQueue[message.Author.Nickname] = append(pmQueue[message.Author.Nickname], embed)
	}

	for username, queue := range pmQueue {
		targetChannelID := service.GetPrivateMessageChannelID(username)
		chunkSliceEmbeds := helper.ChunkSliceEmbeds(queue, 10)
		for _, embeds := range chunkSliceEmbeds {
			bot.ChannelMessageSendComplex(targetChannelID, &discordgo.MessageSend{
				Embeds: embeds,
			})
		}
	}
}

func (service *DiscordServiceImpl) GetPrivateMessageChannelID(username string) string {
	bot := service.GetSession()
	guildID := "1197918994301202573"
	channels, err := bot.GuildChannels(guildID)
	if err != nil {
		log.Println(err)
	}

	pmA := "1199688087752671302"
	pmASize := 0
	pmB := "1199910507621142540"
	pmBSize := 0

	var targetChannelID string
	for _, channel := range channels {
		if channel.ParentID == pmA || channel.ParentID == pmB {
			if channel.ParentID == pmA {
				pmASize++
			}
			if channel.ParentID == pmB {
				pmBSize++
			}

			if channel.Name == strings.ToLower(username) {
				targetChannelID = channel.ID
			}
		}
	}

	if targetChannelID == "" {
		targetRoleID := service.GetPrivateMessageRoleID(username)
		targetParentID := pmA
		if pmASize >= 50 {
			targetParentID = pmB
		}
		ch, _ := bot.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
			Name:     strings.ToLower(username),
			Type:     discordgo.ChannelTypeGuildText,
			Topic:    fmt.Sprintf("JKT48 PM %s | TIDAK BOLEH SCREENSHOT", username),
			ParentID: targetParentID,
			PermissionOverwrites: []*discordgo.PermissionOverwrite{
				{
					ID:    targetRoleID,
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel | discordgo.PermissionReadMessageHistory,
				},
				{
					ID:   "1197918994301202573",
					Type: discordgo.PermissionOverwriteTypeRole,
					Deny: discordgo.PermissionAll,
				},
			},
		})
		targetChannelID = ch.ID
	}
	return targetChannelID
}

func (service *DiscordServiceImpl) GetPrivateMessageRoleID(username string) string {
	bot := service.GetSession()
	guildID := "1197918994301202573"
	roles, _ := bot.GuildRoles(guildID)

	var targetRoleID string
	for _, role := range roles {
		if role.Name == username {
			targetRoleID = role.ID
			break
		}
	}

	if targetRoleID == "" {
		r, _ := bot.GuildRoleCreate(guildID, &discordgo.RoleParams{
			Name: username,
		})
		targetRoleID = r.ID
	}
	return targetRoleID
}

func (service *DiscordServiceImpl) SendRankUpdate(ranks []model.PMRanking, updateType string, byteImage []byte) {
	bot := service.GetSession()

	timezone, _ := time.LoadLocation("Asia/Jakarta")
	yesterday := time.Now().In(timezone).AddDate(0, 0, -1).Format("02/01/2006")

	var title string
	var color int
	var description string

	switch strings.ToLower(updateType) {
	case "daily":
		title = "Daily PM Stats Update"
		color = 0x57F287
		description = fmt.Sprintf("💚 Update harian top 5 Private Messages kemaren ` (%s) `", yesterday)
	case "weekly":
		lastWeek := time.Now().In(timezone).AddDate(0, 0, -8).Format("02/01/2006")
		title = "Weekly PM Stats Update"
		color = 0xE67E22
		description = fmt.Sprintf("🧡 Update mingguan top 10 Private Messages minggu kemaren ` (%s - %s) `", lastWeek, yesterday)
	case "monthly":
		lastMonth := time.Now().In(timezone).AddDate(0, -1, -1)
		title = "Monthly PM Stats Update"
		color = 0x3498DB
		description = fmt.Sprintf("💙 Update bulanan top 10 Private Messages bulan %s ` (%s - %s) `", lastMonth.Month().String(), lastMonth.Format("02/01/2006"), yesterday)
	}

	embed := &discordgo.MessageEmbed{
		URL:         "https://jkt48.safatanc.com",
		Title:       title,
		Description: fmt.Sprintf("%s\n\nMekanisme penghitungan point:\n- Text: 1 point\n- Image: 2 point\n- Voice: 3 point", description),
		Color:       color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "JKT48Lab by safatanc.com",
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://img.jpg",
		},
	}

	bot.ChannelMessageSendComplex("1200958877391400993", &discordgo.MessageSend{
		File: &discordgo.File{
			Name:   "img.jpg",
			Reader: bytes.NewReader(byteImage),
		},
		Embed: embed,
	})
	log.Println("[INFO] Success updating", title)
}
