package utils

import (
	"math/rand"
	"strings"
	"time"
)

type ScrambleUtils struct {
	rng *rand.Rand
}

func NewScrambleUtils() *ScrambleUtils {
	return &ScrambleUtils{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func GenerateScrambleHint(word string) string {
	if len(word) == 0 {
		return ""
	}

	if len(word) <= 2 {
		return word
	}

	if len(word) == 3 {
		return string(word[0]) + "_" + string(word[2])
	}

	hint := string(word[0])
	for i := 1; i < len(word)-1; i++ {
		hint += "_"
	}
	hint += string(word[len(word)-1])

	return hint
}

func ValidateScrambledWord(original, scrambled string) bool {
	return !strings.EqualFold(original, scrambled)
}
