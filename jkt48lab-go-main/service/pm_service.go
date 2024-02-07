package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"html/template"
	"jkt48lab/helper"
	"jkt48lab/model"
	"jkt48lab/repository"
	"log"
	"os"
	"strings"
	"time"
)

type PMService interface {
	FindNewMessages(ctx context.Context) ([]model.PMMessage, error)
	GenerateAccessToken() string
	checkAccessTokenIsValid(accessToken string) bool
	UpdateRanking(ctx context.Context)
	GenerateStatsImage(updateType string, ranks []model.PMRanking, from string, until string) ([]byte, string)
}

type PMServiceImpl struct {
	PMRepository   repository.PMRepository
	DiscordService DiscordService
	TweetService   TweetService
}

func NewPMService(pmRepository repository.PMRepository, discordService DiscordService, tweetService TweetService) PMService {
	return &PMServiceImpl{
		PMRepository:   pmRepository,
		DiscordService: discordService,
		TweetService:   tweetService,
	}
}

func (service *PMServiceImpl) FindNewMessages(ctx context.Context) ([]model.PMMessage, error) {
	accessToken := service.GenerateAccessToken()
	lastMessage := service.PMRepository.FindLastMessage(ctx)
	allMessages, err := service.PMRepository.FindAllMessages(ctx, accessToken)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var foundLastMessage bool
	var newMessages []model.PMMessage
	for _, message := range allMessages {
		birthdayMessage, _ := service.PMRepository.FindBirthdayMessage(ctx, accessToken, message.Author.Nickname)
		if message.Message == birthdayMessage.Message {
			continue
		}
		if message.Id == lastMessage.Id {
			foundLastMessage = true
			break
		}
		if message.CreatedAt != message.UpdatedAt {
			continue
		}
		similarityCheck := strutil.Similarity(birthdayMessage.Message, message.Message, metrics.NewJaccard())
		if similarityCheck < 0.89 {
			newMessages = append([]model.PMMessage{message}, newMessages...)
		}
	}
	if foundLastMessage == false {
		log.Println("[WARNING] Last message not found", lastMessage.Id)
	}
	if len(newMessages) == 0 {
		return nil, nil
	}
	lastMessage = newMessages[len(newMessages)-1]
	service.PMRepository.UpdateLastMessage(ctx, lastMessage)
	return newMessages, nil
}

var AccessToken string

func (service *PMServiceImpl) GenerateAccessToken() string {
	if AccessToken != "" {
		isValid := service.checkAccessTokenIsValid(AccessToken)
		if isValid {
			log.Println("[INFO] Using current accessToken")
			return AccessToken
		} else {
			log.Println("[INFO] Generating new accessToken (expired/invalid)")
		}
	}
	log.Println("[INFO] Generating new accessToken")

	godotenv.Load()

	selectorUsername := `input[name="username"]`
	selectorPassword := `input[name="password"]`
	selectorSubmit := `button[type="submit"]`
	selectorProfile := `div[role="list"]`

	var accessToken string

	tasks := chromedp.Tasks{
		chromedp.Navigate("https://jkt48.primesse.me/signin"),
		chromedp.WaitVisible(selectorUsername, chromedp.ByQuery),
		chromedp.SendKeys(selectorUsername, os.Getenv("PM_EMAIL"), chromedp.ByQuery),
		chromedp.SendKeys(selectorPassword, os.Getenv("PM_PASSWORD"), chromedp.ByQuery),
		chromedp.Click(selectorSubmit, chromedp.ByQuery),
		chromedp.WaitVisible(selectorProfile, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, _ := network.GetCookies().Do(ctx)
			for _, cookie := range cookies {
				if strings.Contains(cookie.Name, "accessToken") {
					accessToken = cookie.Value
				}
			}
			return nil
		}),
		chromedp.Stop(),
	}

	for {
		if accessToken != "" {
			break
		}
		ctx, cancel := chromedp.NewContext(context.Background())
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		err := chromedp.Run(ctx, tasks)
		if err != nil {
			log.Println("[Error] Failed generating accessToken. Retrying...")
			log.Println(err)
		}
		cancel()
	}
	log.Println("[INFO] accessToken generated with size:", len(accessToken))
	AccessToken = accessToken
	return accessToken
}

func (service *PMServiceImpl) checkAccessTokenIsValid(accessToken string) bool {
	resp, err := helper.GraphQLRequest("https://xzqpphzvbzhzvpke6ojjzvbpjq.appsync-api.ap-southeast-1.amazonaws.com/graphql", []byte(`{"query": "a"}`), accessToken)
	if err != nil {
		log.Println(err)
		return false
	}
	if resp.StatusCode == 401 {
		return false
	} else {
		return true
	}
}

func (service *PMServiceImpl) UpdateRanking(ctx context.Context) {
	timezone, _ := time.LoadLocation("Asia/Jakarta")
	scheduler := cron.New(cron.WithLocation(timezone))
	defer scheduler.Stop()
	// Daily at 06:00 WIB | 0 6 * * *
	scheduler.AddFunc("0 6 * * *", func() {
		accessToken := service.GenerateAccessToken()
		today := time.Now().In(timezone)
		yesterday := time.Now().In(timezone).AddDate(0, 0, -1)
		rankings := service.PMRepository.FindRankings(ctx, accessToken, yesterday.Format("2006-01-02"), today.Format("2006-01-02"), 5)
		byteImage, base64img := service.GenerateStatsImage("daily", rankings, "", yesterday.Format("02/01/2006"))
		service.DiscordService.SendRankUpdate(rankings, "daily", byteImage)
		service.TweetService.Post(
			ctx,
			fmt.Sprintf("Selamat pagi, berikut adalah Daily Stats Update JKT48 PM tanggal %s kemaren.", yesterday.Format("02/01/2006")),
			base64img,
		)
	})

	// Weekly at 06:00 WIB (8, 15, 22, 29) | 0 6 */7 * *
	scheduler.AddFunc("0 6 */7 * *", func() {
		accessToken := service.GenerateAccessToken()
		today := time.Now().In(timezone)
		yesterday := time.Now().In(timezone).AddDate(0, 0, -1)
		lastWeek := time.Now().In(timezone).AddDate(0, 0, -8)
		rankings := service.PMRepository.FindRankings(ctx, accessToken, lastWeek.Format("2006-01-02"), today.Format("2006-01-02"), 10)
		byteImage, base64img := service.GenerateStatsImage("daily", rankings, lastWeek.Format("02/01/2006"), yesterday.Format("02/01/2006"))
		service.DiscordService.SendRankUpdate(rankings, "daily", byteImage)
		service.TweetService.Post(
			ctx,
			fmt.Sprintf("Selamat pagi, berikut adalah Weekly Stats Update JKT48 PM satu minggu kemaren (%s - %s).", lastWeek.Format("02/01/2006"), yesterday.Format("02/01/2006")),
			base64img,
		)
	})

	// Monthly at 06:00 WIB | 0 6 1 * *
	scheduler.AddFunc("0 6 1 * *", func() {
		accessToken := service.GenerateAccessToken()
		today := time.Now().In(timezone)
		lastMonth := time.Now().In(timezone).AddDate(0, -1, 0)
		rankings := service.PMRepository.FindRankings(ctx, accessToken, lastMonth.Format("2006-01-02"), today.Format("2006-01-02"), 10)
		byteImage, base64img := service.GenerateStatsImage("monthly", rankings, lastMonth.Month().String(), string(rune(lastMonth.Year())))
		service.DiscordService.SendRankUpdate(rankings, "monthly", byteImage)
		service.TweetService.Post(
			ctx,
			fmt.Sprintf("Selamat pagi, berikut adalah Monthly Stats Update JKT48 PM bulan %s %d kemaren.", lastMonth.Month().String(), lastMonth.Year()),
			base64img,
		)
	})
	go scheduler.Start()
}

func (service *PMServiceImpl) GenerateStatsImage(updateType string, ranks []model.PMRanking, from string, until string) ([]byte, string) {
	tem := template.Must(template.ParseFiles("./templates/pm_stats.html"))
	htmlBuffer := new(bytes.Buffer)
	err := tem.Execute(htmlBuffer, model.PMStatsTemplate{
		Type:  strings.ToTitle(updateType),
		Ranks: ranks,
		From:  from,
		Until: until,
	})
	if err != nil {
		log.Fatal(err)
	}

	cc, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var ssData []byte
	chromedp.Run(cc, chromedp.Tasks{
		chromedp.Navigate(fmt.Sprintf("data:text/html,%s", htmlBuffer.String())),
		chromedp.WaitVisible("section", chromedp.ByQuery),
		chromedp.Screenshot("section", &ssData),
	})
	base64img := base64.StdEncoding.EncodeToString(ssData)
	return ssData, base64img
}
