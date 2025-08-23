package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"twitchgo/types"
)

type InMemoryPointsDB struct {
	users map[string]*types.UserData
	mutex sync.RWMutex
	rng   *rand.Rand
	path  string
}

func NewInMemoryPointsDB() *InMemoryPointsDB {
	db := &InMemoryPointsDB{
		users: make(map[string]*types.UserData),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
		path:  filepath.Join("data", "user_data.json"),
	}

	if err := db.LoadFromFile(); err != nil {
		log.Printf("Failed to load user data: %v", err)
		log.Println("Starting with empty user database")
	}

	return db
}

func (db *InMemoryPointsDB) ValidateUser(username string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	username = strings.ToLower(username)
	if _, exists := db.users[username]; !exists {
		db.users[username] = &types.UserData{
			Username:   username,
			Points:     0,
			GambleLoss: 0,
		}
		log.Printf("Created new user: %s", username)
	}
	return nil
}

func (db *InMemoryPointsDB) GetPoints(username string) int {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	username = strings.ToLower(username)
	if user, exists := db.users[username]; exists {
		return user.Points
	}
	return 0
}

func (db *InMemoryPointsDB) AddPoints(username string, amount int) error {
	if amount < 0 {
		return fmt.Errorf("cannot add negative points")
	}

	db.ValidateUser(username)
	db.mutex.Lock()
	defer db.mutex.Unlock()

	username = strings.ToLower(username)
	db.users[username].Points += amount
	log.Printf("Added %d points to %s (new balance: %d)", amount, username, db.users[username].Points)
	return nil
}

func (db *InMemoryPointsDB) SubtractPoints(username string, amount int) error {
	if amount < 0 {
		return fmt.Errorf("cannot subtract negative points")
	}

	db.ValidateUser(username)
	db.mutex.Lock()
	defer db.mutex.Unlock()

	username = strings.ToLower(username)
	user := db.users[username]

	newBalance := user.Points - amount
	if newBalance < 0 {
		newBalance = 0
		log.Printf("Warning: %s would have negative points, setting to 0", username)
	}

	user.Points = newBalance
	log.Printf("Subtracted %d points from %s (new balance: %d)", amount, username, user.Points)
	return nil
}

func (db *InMemoryPointsDB) GetGambleLoss(username string) int {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	username = strings.ToLower(username)
	if user, exists := db.users[username]; exists {
		return user.GambleLoss
	}
	return 0
}

func (db *InMemoryPointsDB) AddGambleLoss(username string, amount int) error {
	if amount < 0 {
		return fmt.Errorf("cannot add negative gamble loss")
	}

	db.ValidateUser(username)
	db.mutex.Lock()
	defer db.mutex.Unlock()

	username = strings.ToLower(username)
	db.users[username].GambleLoss += amount
	log.Printf("Added %d gamble loss to %s (total loss: %d)", amount, username, db.users[username].GambleLoss)
	return nil
}

func (db *InMemoryPointsDB) TransferPoints(sender, receiver string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	sender = strings.ToLower(sender)
	receiver = strings.ToLower(receiver)

	if sender == receiver {
		return fmt.Errorf("cannot transfer to yourself")
	}

	db.ValidateUser(sender)
	db.ValidateUser(receiver)

	senderPoints := db.GetPoints(sender)
	if senderPoints < amount {
		return fmt.Errorf("insufficient points")
	}

	if err := db.SubtractPoints(sender, amount); err != nil {
		return err
	}
	if err := db.AddPoints(receiver, amount); err != nil {
		// Revert sender subtraction
		db.AddPoints(sender, amount)
		return err
	}

	log.Printf("Transferred %d points from %s to %s", amount, sender, receiver)
	return nil
}

func (db *InMemoryPointsDB) Gamble(username string, wager int, format string, winOdds float64) (string, int, int, error) {
	if winOdds < 0 || winOdds > 1 {
		return "", 0, 0, fmt.Errorf("win odds must be between 0 and 1")
	}

	if format != "points" && format != "percent" && format != "all" {
		return "", 0, 0, fmt.Errorf("format must be 'points', 'percent', or 'all'")
	}

	db.ValidateUser(username)
	currentPoints := db.GetPoints(username)

	if currentPoints == 0 {
		return "no points", currentPoints, 0, nil
	}

	var actualWager int

	switch format {
	case "points":
		if currentPoints < wager {
			return "not enough points", currentPoints, 0, nil
		}
		actualWager = wager

	case "percent":
		if wager > 100 {
			return "invalid percent", currentPoints, 0, nil
		}
		actualWager = int(float64(currentPoints) * float64(wager) / 100.0)
		if actualWager == 0 {
			return "not enough points", currentPoints, 0, nil
		}
		if currentPoints < actualWager {
			return "not enough points", currentPoints, 0, nil
		}

	case "all":
		actualWager = currentPoints
	}

	// Determine outcome
	outcome := "lose"
	if db.rng.Float64() < winOdds {
		outcome = "win"
	}

	var newBalance int
	if outcome == "win" {
		db.AddPoints(username, actualWager)
		newBalance = currentPoints + actualWager
		log.Printf("Gamble result: %s won %d points (balance: %d -> %d)", username, actualWager, currentPoints, newBalance)
		return "win", newBalance, actualWager, nil
	} else {
		db.SubtractPoints(username, actualWager)
		db.AddGambleLoss(username, actualWager)
		newBalance = currentPoints - actualWager
		log.Printf("Gamble result: %s lost %d points (balance: %d -> %d)", username, actualWager, currentPoints, newBalance)
		return "lose", newBalance, actualWager, nil
	}
}

func (db *InMemoryPointsDB) GetTopPoints(limit int) ([]string, []int) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	type userScore struct {
		username string
		points   int
	}

	var users []userScore
	for _, user := range db.users {
		users = append(users, userScore{user.Username, user.Points})
	}

	// Simple bubble sort for small datasets
	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			if users[j].points > users[i].points {
				users[i], users[j] = users[j], users[i]
			}
		}
	}

	var usernames []string
	var points []int

	maxResults := limit
	if len(users) < limit {
		maxResults = len(users)
	}

	for i := 0; i < maxResults; i++ {
		usernames = append(usernames, users[i].username)
		points = append(points, users[i].points)
	}

	return usernames, points
}

func (db *InMemoryPointsDB) GetTopGambleLoss(limit int) ([]string, []int) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	type userScore struct {
		username   string
		gambleLoss int
	}

	var users []userScore
	for _, user := range db.users {
		users = append(users, userScore{user.Username, user.GambleLoss})
	}

	// Simple bubble sort for small datasets
	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			if users[j].gambleLoss > users[i].gambleLoss {
				users[i], users[j] = users[j], users[i]
			}
		}
	}

	var usernames []string
	var losses []int

	maxResults := limit
	if len(users) < limit {
		maxResults = len(users)
	}

	for i := 0; i < maxResults; i++ {
		usernames = append(usernames, users[i].username)
		losses = append(losses, users[i].gambleLoss)
	}

	return usernames, losses
}

func (db *InMemoryPointsDB) GetRank(username string) (int, int) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	username = strings.ToLower(username)

	type userScore struct {
		username   string
		points     int
		gambleLoss int
	}

	var users []userScore
	for _, user := range db.users {
		users = append(users, userScore{user.Username, user.Points, user.GambleLoss})
	}

	// Sort by points (descending)
	pointsRank := 1
	for _, user := range users {
		if user.points > db.users[username].Points {
			pointsRank++
		}
	}

	// Sort by gamble loss (descending)
	lossRank := 1
	for _, user := range users {
		if user.gambleLoss > db.users[username].GambleLoss {
			lossRank++
		}
	}

	return pointsRank, lossRank
}

func (db *InMemoryPointsDB) SaveToFile() error {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var users []types.UserData
	for _, user := range db.users {
		users = append(users, *user)
	}

	file, err := os.Create(db.path)
	if err != nil {
		return fmt.Errorf("failed to create user data file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(users); err != nil {
		return fmt.Errorf("failed to encode user data: %w", err)
	}

	log.Printf("Successfully saved %d users to %s", len(users), db.path)
	return nil
}

func (db *InMemoryPointsDB) LoadFromFile() error {
	file, err := os.Open(db.path)
	if err != nil {
		return fmt.Errorf("failed to open user data file: %w", err)
	}
	defer file.Close()

	var users []types.UserData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&users); err != nil {
		return fmt.Errorf("failed to decode user data: %w", err)
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.users = make(map[string]*types.UserData)
	for _, user := range users {
		db.users[strings.ToLower(user.Username)] = &types.UserData{
			Username:   user.Username,
			Points:     user.Points,
			GambleLoss: user.GambleLoss,
		}
	}

	log.Printf("Successfully loaded %d users from %s", len(users), db.path)
	return nil
}
