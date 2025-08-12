package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"twitchgo/types"
)

type InMemoryScrambleDB struct {
	words []types.ScrambleWord
	rng   *rand.Rand
}

func NewInMemoryScrambleDB() *InMemoryScrambleDB {
	db := &InMemoryScrambleDB{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if err := db.ReloadWords(); err != nil {
		log.Printf("Failed to load scramble words: %v", err)
	}

	return db
}

func (db *InMemoryScrambleDB) ReloadWords() error {
	dataPath := filepath.Join("data", "scramble_words.json")

	file, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("failed to open scramble words file: %w", err)
	}
	defer file.Close()

	var words []types.ScrambleWord
	if err := json.NewDecoder(file).Decode(&words); err != nil {
		return fmt.Errorf("failed to decode scramble words: %w", err)
	}

	enabledWords := make([]types.ScrambleWord, 0)
	for _, word := range words {
		if word.Enabled {
			enabledWords = append(enabledWords, word)
		}
	}

	db.words = enabledWords
	log.Printf("Loaded %d scramble words", len(db.words))

	return nil
}

func (db *InMemoryScrambleDB) GetRandomWord() *types.ScrambleWord {
	if len(db.words) == 0 {
		return nil
	}

	index := db.rng.Intn(len(db.words))
	word := db.words[index]
	return &word
}

func ScrambleString(s string) string {
	if len(s) <= 1 {
		return s
	}

	runes := []rune(s)
	shuffled := string(runes)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for shuffled == s {
		for i := len(runes) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			runes[i], runes[j] = runes[j], runes[i]
		}
		shuffled = string(runes)
	}

	return shuffled
}

func CheckScrambleGuess(guess, originalWord string) (bool, float64) {
	guess = SanitizeMessage(guess)
	originalWord = SanitizeMessage(originalWord)

	if guess == originalWord {
		return true, 1.0
	}

	if strings.Contains(guess, originalWord) {
		return true, 1.0
	}

	similarity := CalculateSimilarity(guess, originalWord)

	if similarity >= 0.85 {
		return true, similarity
	}

	return false, similarity
}
