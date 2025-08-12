package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"twitchgo/types"
	"twitchgo/utils"

	"github.com/gempir/go-twitch-irc/v4"
)

type ScrambleGame struct {
	Active        bool
	Word          types.ScrambleWord
	ScrambledWord string
	StartTime     time.Time
	HintGiven     bool
	LastStarted   time.Time
	ctx           context.Context
	cancel        context.CancelFunc
}

type ScrambleManager struct {
	game       *ScrambleGame
	database   types.ScrambleDatabase
	config     ScrambleConfig
	messageGen ScrambleMessageGenerator
}

type ScrambleConfig struct {
	Cooldown  time.Duration
	HintTime  time.Duration
	Timeout   time.Duration
	MaxLength int
}

type ScrambleMessageGenerator interface {
	FormatScramble(scrambledWord string) string
	FormatCorrectAnswer(user, answer string, points int) string
	FormatCloseAnswer(user, guess string, similarity float64) string
	FormatHint(hint string) string
	FormatTimeout(answer string) string
	FormatAlreadyRunning(user string) string
	FormatStopped() string
	FormatNoWords() string
}

type defaultScrambleMessageGenerator struct{}

func (g *defaultScrambleMessageGenerator) FormatScramble(scrambledWord string) string {
	return fmt.Sprintf("[Embaralha] Desembaralhe esta palavra: %s 🧩", scrambledWord)
}

func (g *defaultScrambleMessageGenerator) FormatCorrectAnswer(user, answer string, points int) string {
	return fmt.Sprintf("[Embaralha] @%s Parabéns! Você acertou e ganhou %d pontos! Gayge Clap A palavra era: \"%s\"",
		user, points, answer)
}

func (g *defaultScrambleMessageGenerator) FormatCloseAnswer(user, guess string, similarity float64) string {
	return fmt.Sprintf("[Embaralha] @%s \"%s\" está perto! [Similaridade %.0f%%]",
		user, guess, similarity*100)
}

func (g *defaultScrambleMessageGenerator) FormatHint(hint string) string {
	return fmt.Sprintf("[Embaralha] Dica: %s", hint)
}

func (g *defaultScrambleMessageGenerator) FormatTimeout(answer string) string {
	return fmt.Sprintf("[Embaralha] Tempo esgotado! Madge A palavra era: %s", answer)
}

func (g *defaultScrambleMessageGenerator) FormatAlreadyRunning(user string) string {
	return fmt.Sprintf("[Embaralha] @%s Scramble já está em andamento.", user)
}

func (g *defaultScrambleMessageGenerator) FormatStopped() string {
	return "[Embaralha] MrDestructoid Scramble parou."
}

func (g *defaultScrambleMessageGenerator) FormatNoWords() string {
	return "[Embaralha] Nenhuma palavra disponível."
}

func NewScrambleManager(database types.ScrambleDatabase, config ScrambleConfig) *ScrambleManager {
	return &ScrambleManager{
		game:       &ScrambleGame{},
		database:   database,
		config:     config,
		messageGen: &defaultScrambleMessageGenerator{},
	}
}

func (sm *ScrambleManager) StartScramble(client *twitch.Client, message twitch.PrivateMessage) {
	if sm.game.Active {
		if time.Since(sm.game.StartTime) > 5*time.Second {
			client.Say(message.Channel, sm.messageGen.FormatAlreadyRunning(message.User.DisplayName))
		}
		return
	}

	if time.Since(sm.game.LastStarted) < sm.config.Cooldown {
		log.Printf("Scramble command blocked -- in silent cooldown.")
		return
	}

	word := sm.database.GetRandomWord()
	if word == nil {
		client.Say(message.Channel, sm.messageGen.FormatNoWords())
		return
	}

	scrambledWord := utils.ScrambleString(word.Word)

	// Ensure the scrambled word is actually different from the original
	if !utils.ValidateScrambledWord(word.Word, scrambledWord) {
		// If still the same, try again
		scrambledWord = utils.ScrambleString(word.Word)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sm.config.Timeout)
	sm.game = &ScrambleGame{
		Active:        true,
		Word:          *word,
		ScrambledWord: scrambledWord,
		StartTime:     time.Now(),
		LastStarted:   time.Now(),
		HintGiven:     false,
		ctx:           ctx,
		cancel:        cancel,
	}

	client.Say(message.Channel, sm.messageGen.FormatScramble(scrambledWord))

	log.Printf("Scramble Word: %s", word.Word)
	log.Printf("Scrambled: %s", scrambledWord)
	log.Printf("Scramble ID: %s", word.ID)

	go sm.manageTimer(client, message.Channel)
}

func (sm *ScrambleManager) StopScramble(client *twitch.Client, message twitch.PrivateMessage) {
	if sm.game.Active {
		sm.stopGame()
		client.Say(message.Channel, sm.messageGen.FormatStopped())
		log.Println("Scramble stopped by moderator")
	}
}

func (sm *ScrambleManager) CheckAnswer(client *twitch.Client, message twitch.PrivateMessage, checkFunc func(string, string) (bool, float64)) {
	if !sm.game.Active || len(message.Message) > sm.config.MaxLength {
		return
	}

	correct, similarity := checkFunc(message.Message, sm.game.Word.Word)

	if correct {
		sm.handleCorrectAnswer(client, message, similarity)
	} else if similarity >= 0.75 && len(message.Message) < sm.config.MaxLength {
		sm.handleCloseAnswer(client, message, similarity)
	}
}

func (sm *ScrambleManager) IsActive() bool {
	return sm.game.Active
}

func (sm *ScrambleManager) GetCurrentWord() *types.ScrambleWord {
	if sm.game.Active {
		return &sm.game.Word
	}
	return nil
}

func (sm *ScrambleManager) manageTimer(client *twitch.Client, channel string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for sm.game.Active {
		select {
		case <-sm.game.ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(sm.game.StartTime)

			if elapsed >= sm.config.HintTime && !sm.game.HintGiven && sm.game.Active {
				sm.giveHint(client, channel)
			}

			if elapsed >= sm.config.Timeout && sm.game.Active {
				sm.handleTimeout(client, channel)
				return
			}
		}
	}
}

func (sm *ScrambleManager) giveHint(client *twitch.Client, channel string) {
	sm.game.HintGiven = true
	hint := utils.GenerateScrambleHint(sm.game.Word.Word)
	client.Say(channel, sm.messageGen.FormatHint(hint))
}

func (sm *ScrambleManager) handleTimeout(client *twitch.Client, channel string) {
	sm.stopGame()
	client.Say(channel, sm.messageGen.FormatTimeout(sm.game.Word.Word))
	log.Printf("Scramble timeout - Answer was: %s", sm.game.Word.Word)
}

func (sm *ScrambleManager) handleCorrectAnswer(client *twitch.Client, message twitch.PrivateMessage, similarity float64) {
	sm.stopGame()

	points := 6
	if similarity >= 0.95 {
		points = 8
	}

	client.Say(message.Channel, sm.messageGen.FormatCorrectAnswer(
		message.User.DisplayName, sm.game.Word.Word, points))

	log.Printf("[Scramble] %s answered correctly with similarity %.2f",
		message.User.DisplayName, similarity)
}

func (sm *ScrambleManager) handleCloseAnswer(client *twitch.Client, message twitch.PrivateMessage, similarity float64) {
	client.Say(message.Channel, sm.messageGen.FormatCloseAnswer(
		message.User.DisplayName, message.Message, similarity))
	log.Printf("[Scramble] %s is close (%.0f%%)", message.User.DisplayName, similarity*100)
}

func (sm *ScrambleManager) stopGame() {
	if sm.game.cancel != nil {
		sm.game.cancel()
	}
	sm.game.Active = false
}
