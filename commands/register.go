package commands

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type CommandFunc func(client *twitch.Client, message twitch.PrivateMessage)

var commandMap = map[string]CommandFunc{
	"hora":     Time,
	"bot":      Hello,
	"paraquiz": StopTrivia,
}

func Handle(client *twitch.Client, message twitch.PrivateMessage, prefix string) {
	cmd := strings.ToLower(strings.TrimPrefix(message.Message, prefix))

	if handler, ok := commandMap[cmd]; ok {
		handler(client, message)
	}
}
