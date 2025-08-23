package commands

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type CommandFunc func(client *twitch.Client, message twitch.PrivateMessage)

var commandMap = map[string]CommandFunc{
	"hora":          Time,
	"bot":           Hello,
	"quiz":          Trivia,
	"paraquiz":      StopTrivia,
	"embaralha":     Scramble,
	"paraembaralha": StopScramble,
	"roleta":        Roulette,
	"pontos":        Points,
	"dar":           GivePoints,
	"doar":          GivePoints,
	"enviar":        GivePoints,
	"top":           TopPoints,
	"toppontos":     TopPoints,
	"topperda":      TopGambleLoss,
	"rank":          Rank,
	"ranking":       Rank,
	"addpontos":     AddPointsCommand,
	"diario":        DailyPoints,
}

func Handle(client *twitch.Client, message twitch.PrivateMessage, prefix string) {
	cmd := strings.ToLower(strings.TrimPrefix(message.Message, prefix))

	if handler, ok := commandMap[cmd]; ok {
		handler(client, message)
	}
}
