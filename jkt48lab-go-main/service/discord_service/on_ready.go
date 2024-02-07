package discord_service

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

func OnReady(bot *discordgo.Session, ready *discordgo.Ready) {
	bot.UpdateListeningStatus("JKT48")
	log.Println(fmt.Sprintf("[Running] Discord BOT %s", ready.User.Username))
}
