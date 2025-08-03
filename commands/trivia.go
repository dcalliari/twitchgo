package commands

import (
	"time"

	"twitchgo/service"
	"twitchgo/utils"

	"github.com/gempir/go-twitch-irc/v4"
)

var triviaManager *service.TriviaManager

func init() {
	database := utils.NewInMemoryTriviaDB()

	config := service.TriviaConfig{
		Cooldown:  10 * time.Second,
		HintTime:  20 * time.Second,
		Timeout:   30 * time.Second,
		MaxLength: 250,
	}

	triviaManager = service.NewTriviaManager(database, config)
}

func Trivia(client *twitch.Client, message twitch.PrivateMessage) {
	triviaManager.StartTrivia(client, message)
}

func StopTrivia(client *twitch.Client, message twitch.PrivateMessage) {
	triviaManager.StopTrivia(client, message)
}

func CheckTriviaAnswer(client *twitch.Client, message twitch.PrivateMessage) {
	triviaManager.CheckAnswer(client, message, utils.CheckTriviaGuess)
}
