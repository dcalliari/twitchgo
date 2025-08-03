package commands

import (
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

func Time(client *twitch.Client, message twitch.PrivateMessage) {
	now := time.Now().Format("15:04:05")
	client.Say(message.Channel, "🕒 Agora são "+now)
}
