package main

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"jkt48lab/model"
	"jkt48lab/repository"
	"jkt48lab/service"
	"jkt48lab/service/discord_service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	discordService := service.NewDiscordService()
	tweetService := service.NewTweetService()
	srLiveRepository := repository.NewSRLiveRepository()
	srLiveService := service.NewSRLiveService(srLiveRepository, discordService)
	idnLiveRepository := repository.NewIDNLiveRepository()
	idnLiveService := service.NewIDNLiveService(idnLiveRepository)
	pmRepository := repository.NewPMRepository()
	pmService := service.NewPMService(pmRepository, discordService, tweetService)
	ctx := context.Background()

	log.Println("[Running] JKT48Lab")

	bot := discordService.GetSession()
	bot.AddHandler(discord_service.OnReady)
	bot.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages
	err := bot.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()
	var onLives model.OnLives

	go func() {
		log.Println("[Watching] Private Messages")
		for {
			messages, _ := pmService.FindNewMessages(ctx)
			if len(messages) > 0 {
				discordService.SendPrivateMessages(messages)
				//for _, message := range messages {
				//	log.Println(fmt.Sprintf("%s | %s", message.Author.Nickname, message.Message))
				//}
			}
			time.Sleep(2 * time.Minute)
		}
	}()

	go func() {
		pmService.UpdateRanking(ctx)
	}()

	go func() {
		log.Println("[Watching] Showroom Lives")
		for {
			srLives, err := srLiveService.FindAll(ctx)
			if err != nil {
				log.Println(err)
			} else {
				for _, live := range srLives {
					srLiveService.SendNotification(ctx, &onLives, live)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		log.Println("[Watching] IDN Lives")
		for {
			idnLives, err := idnLiveService.FindAll(ctx)
			if err != nil {
				log.Println(err)
			} else {
				for _, live := range idnLives {
					idnLiveService.SendNotification(ctx, &onLives, live)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	log.Println("[Running] Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
