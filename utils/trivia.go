package utils

import (
	"strings"
	"time"
)

func GenerateHint(answer string) string {
	if len(answer) == 0 {
		return ""
	}

	hint := strings.Repeat("_", len(answer))
	runes := []rune(hint)
	answerRunes := []rune(answer)

	for i, char := range answerRunes {
		if char == ' ' {
			runes[i] = ' '
		}
	}

	revealCount := 0
	switch {
	case len(answer) <= 2:
		revealCount = 0
	case len(answer) == 3:
		revealCount = 1
	case len(answer) == 4:
		revealCount = 2
	case len(answer) < 10:
		revealCount = 3
	default:
		revealCount = 6
	}

	revealed := 0
	for i := 0; i < len(answerRunes) && revealed < revealCount; i++ {
		if answerRunes[i] != ' ' {
			runes[i] = answerRunes[i]
			revealed++
		}
	}

	result := string(runes)
	if len(answer) > 10 {
		result = result[:10] + "..."
	}

	return result
}

func CheckTriviaGuess(guess, answer string) (bool, float64) {
	guess = SanitizeMessage(guess)
	answer = SanitizeMessage(answer)

	if guess == answer {
		return true, 1.0
	}

	if strings.Contains(guess, answer) {
		return true, 1.0
	}

	similarity := CalculateSimilarity(guess, answer)

	if similarity >= 0.92 {
		return true, similarity
	} else if similarity >= 0.75 {
		return true, similarity
	}

	return false, similarity
}

func CalculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	longer := s1
	shorter := s2
	if len(s1) < len(s2) {
		longer = s2
		shorter = s1
	}

	if len(longer) == 0 {
		return 1.0
	}

	matches := 0
	for i, char := range shorter {
		if i < len(longer) && rune(longer[i]) == char {
			matches++
		}
	}

	subseqMatches := longestCommonSubsequence(s1, s2)

	positionScore := float64(matches) / float64(len(longer))
	subseqScore := float64(subseqMatches) / float64(len(longer))

	if positionScore > subseqScore {
		return positionScore
	}
	return subseqScore
}

func longestCommonSubsequence(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m == 0 || n == 0 {
		return 0
	}

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func GetRandomSeed() int64 {
	return time.Now().UnixNano()
}
