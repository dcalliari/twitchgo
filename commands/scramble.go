package commands

import (
	"time"

	"twitchgo/service"
	"twitchgo/utils"

	"github.com/gempir/go-twitch-irc/v4"
)

var scrambleManager *service.ScrambleManager

func init() {
	database := utils.NewInMemoryScrambleDB()

	config := service.ScrambleConfig{
		Cooldown:  10 * time.Second,
		HintTime:  20 * time.Second,
		Timeout:   40 * time.Second,
		MaxLength: 250,
	}

	scrambleManager = service.NewScrambleManager(database, config)
}

func Scramble(client *twitch.Client, message twitch.PrivateMessage) {
	scrambleManager.StartScramble(client, message)
}

func StopScramble(client *twitch.Client, message twitch.PrivateMessage) {
	scrambleManager.StopScramble(client, message)
}

func CheckScrambleAnswer(client *twitch.Client, message twitch.PrivateMessage) {
	scrambleManager.CheckAnswer(client, message, utils.CheckScrambleGuess)
}
