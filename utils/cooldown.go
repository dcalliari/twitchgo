package utils

import (
	"sync"
	"time"
)

var cooldowns = make(map[string]time.Time)
var mu sync.Mutex

func IsOnCooldown(user, cmd string, duration time.Duration) bool {
	mu.Lock()
	defer mu.Unlock()

	key := user + ":" + cmd
	now := time.Now()

	if t, exists := cooldowns[key]; exists && now.Before(t) {
		return true
	}

	cooldowns[key] = now.Add(duration)
	return false
}
