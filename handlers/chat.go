package handlers

import (
	"log"
	"strings"

	"twitchgo/commands"

	"github.com/gempir/go-twitch-irc/v4"
)

func OnMessage(client *twitch.Client, message twitch.PrivateMessage, prefix string) {
	log.Printf("[%s]: %s", message.User.Name, message.Message)

	if strings.HasPrefix(message.Message, prefix) {
		commands.Handle(client, message, prefix)
		return
	}

	commands.CheckTriviaAnswer(client, message)
	commands.CheckScrambleAnswer(client, message)

	if strings.Contains(strings.ToLower(message.Message), "bot") {
		client.Say(message.Channel, "👀 Chamou?")
	}
}
