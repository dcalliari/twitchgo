package commands

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"twitchgo/types"
	"twitchgo/utils"

	"github.com/gempir/go-twitch-irc/v4"
)

var pointsDB types.PointsDatabase

func init() {
	pointsDB = utils.NewInMemoryPointsDB()
}

func Roulette(client *twitch.Client, message twitch.PrivateMessage) {
	if utils.IsOnCooldown("global", "roulette", 5*time.Second) {
		log.Println("Roulette command blocked -- in silent cooldown.")
		return
	}

	parts := strings.Fields(message.Message)
	if len(parts) < 2 {
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Awkward Por favor especifique uma aposta.", message.User.DisplayName))
		return
	}

	wagerStr := strings.TrimSpace(parts[1])
	var wager int
	var format string
	var err error

	if wagerStr == "all" {
		format = "all"
		wager = 0
	} else if strings.HasSuffix(wagerStr, "%") {
		format = "percent"
		percentStr := strings.TrimSuffix(wagerStr, "%")
		wager, err = strconv.Atoi(percentStr)
		if err != nil {
			log.Printf("Invalid wager format: %s", wagerStr)
			return
		}
		if wager > 100 {
			client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Weirdge Você não pode apostar mais de 100%% dos seus pontos.", message.User.DisplayName))
			return
		}
	} else {
		format = "points"
		wager, err = strconv.Atoi(wagerStr)
		if err != nil {
			log.Printf("Invalid wager format: %s", wagerStr)
			return
		}
	}

	if format != "all" && wager < 0 {
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Madgay A aposta deve ser positiva.", message.User.DisplayName))
		return
	}

	if format != "all" && wager == 0 {
		client.Say(message.Channel, fmt.Sprintf("[Roleta] 🫵 ICANT @%s acabou de tentar apostar 0 pontos", message.User.DisplayName))
		return
	}

	outcome, newBalance, delta, err := pointsDB.Gamble(message.User.Name, wager, format, 0.50)
	if err != nil {
		log.Printf("Error in gamble function: %v", err)
		return
	}

	switch outcome {
	case "win":
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Gayge Clap Você ganhou %d pontos e agora tem %d pontos.",
			message.User.DisplayName, delta, newBalance))

	case "lose":
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Sadgay SmokeTime Você perdeu %d pontos e agora tem %d pontos.",
			message.User.DisplayName, delta, newBalance))

	case "not enough points":
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Sadgay Você não tem pontos suficientes para isso. Aumente seu dinheiro.",
			message.User.DisplayName))

	case "no points":
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Madgay Você não tem nenhum ponto.",
			message.User.DisplayName))

	case "invalid percent":
		client.Say(message.Channel, fmt.Sprintf("[Roleta] @%s Weirdge Percentual inválido.",
			message.User.DisplayName))
	}
}

func Points(client *twitch.Client, message twitch.PrivateMessage) {
	username := message.User.Name
	points := pointsDB.GetPoints(username)

	client.Say(message.Channel, fmt.Sprintf("@%s Você tem %d pontos.", message.User.DisplayName, points))
}

func GivePoints(client *twitch.Client, message twitch.PrivateMessage) {
	parts := strings.Fields(message.Message)
	if len(parts) != 3 {
		client.Say(message.Channel, fmt.Sprintf("[Doar] @%s Uso: #doar <usuario> <quantia>", message.User.DisplayName))
		return
	}

	receiver := strings.TrimPrefix(parts[1], "@")
	amountStr := parts[2]

	if !isAlphanumeric(receiver) {
		log.Printf("Invalid recipient: %s (not alphanumeric)", receiver)
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		log.Printf("Invalid amount: %s", amountStr)
		return
	}

	if amount <= 0 {
		client.Say(message.Channel, fmt.Sprintf("[Doar] @%s A quantia deve ser positiva.", message.User.DisplayName))
		return
	}

	senderPoints := pointsDB.GetPoints(message.User.Name)
	if senderPoints < amount {
		client.Say(message.Channel, fmt.Sprintf("[Doar] @%s Madgay Você não pode doar mais pontos do que tem.", message.User.DisplayName))
		return
	}

	err = pointsDB.TransferPoints(message.User.Name, receiver, amount)
	if err != nil {
		if strings.Contains(err.Error(), "cannot transfer to yourself") {
			client.Say(message.Channel, fmt.Sprintf("🫵 ICANT @%s Não funcionou.", message.User.DisplayName))
		} else {
			client.Say(message.Channel, fmt.Sprintf("[Doar] @%s Transferência falhou.", message.User.DisplayName))
		}
		return
	}

	client.Say(message.Channel, fmt.Sprintf("[Doar] @%s Doou %d pontos para %s.",
		message.User.DisplayName, amount, receiver))
}

func TopPoints(client *twitch.Client, message twitch.PrivateMessage) {
	usernames, points := pointsDB.GetTopPoints(5)

	if len(usernames) == 0 {
		client.Say(message.Channel, "[TopPontos] Nenhum usuário encontrado.")
		return
	}

	var leaderboard strings.Builder
	leaderboard.WriteString("[TopPontos] Top Points: ")

	for i, username := range usernames {
		if i > 0 {
			leaderboard.WriteString(", ")
		}
		leaderboard.WriteString(fmt.Sprintf("%d. %s (%d)", i+1, username, points[i]))
	}

	client.Say(message.Channel, leaderboard.String())
}

func TopGambleLoss(client *twitch.Client, message twitch.PrivateMessage) {
	usernames, losses := pointsDB.GetTopGambleLoss(5)

	if len(usernames) == 0 {
		client.Say(message.Channel, "[TopPontos] Nenhum usuário encontrado.")
		return
	}

	var leaderboard strings.Builder
	leaderboard.WriteString("[TopPontos] Top Perdas em Apostas: ")

	for i, username := range usernames {
		if i > 0 {
			leaderboard.WriteString(", ")
		}
		leaderboard.WriteString(fmt.Sprintf("%d. %s (%d)", i+1, username, losses[i]))
	}

	client.Say(message.Channel, leaderboard.String())
}

func Rank(client *twitch.Client, message twitch.PrivateMessage) {
	pointsRank, lossRank := pointsDB.GetRank(message.User.Name)

	client.Say(message.Channel, fmt.Sprintf("@%s Sua posição em pontos é %d e sua posição em perdas de apostas é %d.",
		message.User.DisplayName, pointsRank, lossRank))
}

func AddPointsCommand(client *twitch.Client, message twitch.PrivateMessage) {
	parts := strings.Fields(message.Message)
	if len(parts) != 3 {
		client.Say(message.Channel, fmt.Sprintf("[AddPontos] @%s Uso: #addpontos <usuario> <quantia>", message.User.DisplayName))
		return
	}

	targetUser := strings.TrimPrefix(parts[1], "@")
	amountStr := parts[2]

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		client.Say(message.Channel, fmt.Sprintf("[AddPontos] @%s Quantia inválida.", message.User.DisplayName))
		return
	}

	if amount <= 0 {
		client.Say(message.Channel, fmt.Sprintf("[AddPontos] @%s A quantia deve ser positiva.", message.User.DisplayName))
		return
	}

	err = pointsDB.AddPoints(targetUser, amount)
	if err != nil {
		client.Say(message.Channel, fmt.Sprintf("[AddPontos] @%s Erro ao adicionar pontos.", message.User.DisplayName))
		return
	}

	newBalance := pointsDB.GetPoints(targetUser)
	client.Say(message.Channel, fmt.Sprintf("[AddPontos] @%s Adicionou %d pontos a %s (novo saldo: %d).",
		message.User.DisplayName, amount, targetUser, newBalance))
}

func SavePointsData() error {
	return pointsDB.SaveToFile()
}

func DailyPoints(client *twitch.Client, message twitch.PrivateMessage) {
	username := message.User.Name

	dailyAmount := 50
	err := pointsDB.AddPoints(username, dailyAmount)
	if err != nil {
		log.Printf("Error adding daily points to %s: %v", username, err)
		return
	}

	newBalance := pointsDB.GetPoints(username)
	client.Say(message.Channel, fmt.Sprintf("[Diário] @%s Você recebeu %d pontos diários! Novo saldo: %d",
		message.User.DisplayName, dailyAmount, newBalance))
}

func isAlphanumeric(s string) bool {
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return len(s) > 0
}
