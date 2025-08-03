package commands

import (
	"github.com/gempir/go-twitch-irc/v4"
)

func Hello(client *twitch.Client, message twitch.PrivateMessage) {
	client.Say(message.Channel, "🤖 Olá! Eu sou um bot feito em Golang.")
}
