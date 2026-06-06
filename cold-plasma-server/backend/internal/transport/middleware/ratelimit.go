package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket — состояние лимита для одного ключа (обычно IP).
type tokenBucket struct {
	tokens float64
	last   time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rate    float64 // токенов в секунду (скорость восстановления)
	burst   float64 // максимальный запас токенов
}

func newRateLimiter(ratePerSec float64, burst int) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    ratePerSec,
		burst:   float64(burst),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		// Первый запрос: полный запас минус текущий токен.
		rl.buckets[key] = &tokenBucket{tokens: rl.burst - 1, last: now}
		return true
	}

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * rl.rate
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.last = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// cleanupLoop удаляет давно неактивные бакеты, чтобы карта не росла бесконечно.
func (rl *rateLimiter) cleanupLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-30 * time.Minute)
		rl.mu.Lock()
		for k, b := range rl.buckets {
			if b.last.Before(cutoff) {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit возвращает middleware, ограничивающее частоту запросов по IP клиента.
//   - ratePerSec — устойчивая скорость (токенов в секунду);
//   - burst — допустимый всплеск (размер ведра).
//
// Пример: RateLimit(0.2, 10) ≈ 1 запрос каждые 5 секунд, всплеск до 10.
func RateLimit(ratePerSec float64, burst int) gin.HandlerFunc {
	rl := newRateLimiter(ratePerSec, burst)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP(), time.Now()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"message": "Слишком много запросов, попробуйте позже"},
			})
			return
		}
		c.Next()
	}
}
