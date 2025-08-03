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

type TriviaGame struct {
	Active      bool
	Question    types.TriviaQuestion
	StartTime   time.Time
	HintGiven   bool
	LastStarted time.Time
	ctx         context.Context
	cancel      context.CancelFunc
}

type TriviaManager struct {
	game       *TriviaGame
	database   types.TriviaDatabase
	config     TriviaConfig
	messageGen MessageGenerator
}

type TriviaConfig struct {
	Cooldown  time.Duration
	HintTime  time.Duration
	Timeout   time.Duration
	MaxLength int
}

type MessageGenerator interface {
	FormatQuestion(question string) string
	FormatCorrectAnswer(user, answer string, points int) string
	FormatCloseAnswer(user, guess string, similarity float64) string
	FormatHint(hint string) string
	FormatTimeout(answer string) string
	FormatAlreadyRunning(user string) string
	FormatStopped() string
	FormatNoQuestions() string
}

type defaultMessageGenerator struct{}

func (g *defaultMessageGenerator) FormatQuestion(question string) string {
	return fmt.Sprintf("Chatting [Quiz] %s Gayge Clap", question)
}

func (g *defaultMessageGenerator) FormatCorrectAnswer(user, answer string, points int) string {
	return fmt.Sprintf("[Quiz] @%s Você respondeu à pergunta corretamente e ganhou %d pontos. Gayge TeaTime A resposta era: \"%s\"",
		user, points, answer)
}

func (g *defaultMessageGenerator) FormatCloseAnswer(user, guess string, similarity float64) string {
	return fmt.Sprintf("[Quiz] @%s %s está perto. [Similaridade %.0f%%]",
		user, guess, similarity*100)
}

func (g *defaultMessageGenerator) FormatHint(hint string) string {
	return fmt.Sprintf("[Quiz] Dica: %s", hint)
}

func (g *defaultMessageGenerator) FormatTimeout(answer string) string {
	return fmt.Sprintf("[Quiz] Ninguém respondeu corretamente. Madge A resposta era: %s", answer)
}

func (g *defaultMessageGenerator) FormatAlreadyRunning(user string) string {
	return fmt.Sprintf("[Quiz] @%s Quiz já está em andamento.", user)
}

func (g *defaultMessageGenerator) FormatStopped() string {
	return "[Quiz] MrDestructoid Quiz parou."
}

func (g *defaultMessageGenerator) FormatNoQuestions() string {
	return "[Quiz] Nenhuma pergunta disponível."
}

func NewTriviaManager(database types.TriviaDatabase, config TriviaConfig) *TriviaManager {
	return &TriviaManager{
		game:       &TriviaGame{},
		database:   database,
		config:     config,
		messageGen: &defaultMessageGenerator{},
	}
}

func (tm *TriviaManager) StartTrivia(client *twitch.Client, message twitch.PrivateMessage) {
	if tm.game.Active {
		if time.Since(tm.game.StartTime) > 5*time.Second {
			client.Say(message.Channel, tm.messageGen.FormatAlreadyRunning(message.User.DisplayName))
		}
		return
	}

	if time.Since(tm.game.LastStarted) < tm.config.Cooldown {
		log.Printf("Trivia command blocked -- in silent cooldown.")
		return
	}

	question := tm.database.GetRandomQuestion()
	if question == nil {
		client.Say(message.Channel, tm.messageGen.FormatNoQuestions())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), tm.config.Timeout)
	tm.game = &TriviaGame{
		Active:      true,
		Question:    *question,
		StartTime:   time.Now(),
		LastStarted: time.Now(),
		HintGiven:   false,
		ctx:         ctx,
		cancel:      cancel,
	}

	client.Say(message.Channel, tm.messageGen.FormatQuestion(question.Question))

	log.Printf("Trivia Question: %s", question.Question)
	log.Printf("Trivia Answer: %s", question.Answer)
	log.Printf("Trivia QID: %s", question.ID)

	go tm.manageTimer(client, message.Channel)
}

func (tm *TriviaManager) StopTrivia(client *twitch.Client, message twitch.PrivateMessage) {
	if tm.game.Active {
		tm.stopGame()
		client.Say(message.Channel, tm.messageGen.FormatStopped())
		log.Println("Trivia stopped by moderator")
	}
}

func (tm *TriviaManager) CheckAnswer(client *twitch.Client, message twitch.PrivateMessage, checkFunc func(string, string) (bool, float64)) {
	if !tm.game.Active || len(message.Message) > tm.config.MaxLength {
		return
	}

	correct, similarity := checkFunc(message.Message, tm.game.Question.Answer)

	if correct {
		tm.handleCorrectAnswer(client, message, similarity)
	} else if similarity >= 0.82 && len(message.Message) < tm.config.MaxLength {
		tm.handleCloseAnswer(client, message, similarity)
	}
}

func (tm *TriviaManager) IsActive() bool {
	return tm.game.Active
}

func (tm *TriviaManager) GetCurrentQuestion() *types.TriviaQuestion {
	if tm.game.Active {
		return &tm.game.Question
	}
	return nil
}

func (tm *TriviaManager) manageTimer(client *twitch.Client, channel string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for tm.game.Active {
		select {
		case <-tm.game.ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(tm.game.StartTime)

			if elapsed >= tm.config.HintTime && !tm.game.HintGiven && tm.game.Active {
				tm.giveHint(client, channel)
			}

			if elapsed >= tm.config.Timeout && tm.game.Active {
				tm.handleTimeout(client, channel)
				return
			}
		}
	}
}

func (tm *TriviaManager) giveHint(client *twitch.Client, channel string) {
	tm.game.HintGiven = true
	hint := utils.GenerateHint(tm.game.Question.Answer)
	client.Say(channel, tm.messageGen.FormatHint(hint))
}

func (tm *TriviaManager) handleTimeout(client *twitch.Client, channel string) {
	tm.stopGame()
	client.Say(channel, tm.messageGen.FormatTimeout(tm.game.Question.Answer))
	log.Printf("Trivia timeout - Answer was: %s", tm.game.Question.Answer)
}

func (tm *TriviaManager) handleCorrectAnswer(client *twitch.Client, message twitch.PrivateMessage, similarity float64) {
	tm.stopGame()

	points := 8
	if similarity >= 0.92 {
		points = 10
	}

	client.Say(message.Channel, tm.messageGen.FormatCorrectAnswer(
		message.User.DisplayName, tm.game.Question.Answer, points))

	log.Printf("[Trivia] %s answered correctly with similarity %.2f",
		message.User.DisplayName, similarity)
}

func (tm *TriviaManager) handleCloseAnswer(client *twitch.Client, message twitch.PrivateMessage, similarity float64) {
	client.Say(message.Channel, tm.messageGen.FormatCloseAnswer(
		message.User.DisplayName, message.Message, similarity))
	log.Printf("[Trivia] %s is close (%.0f%%)", message.User.DisplayName, similarity*100)
}

func (tm *TriviaManager) stopGame() {
	if tm.game.cancel != nil {
		tm.game.cancel()
	}
	tm.game.Active = false
}
