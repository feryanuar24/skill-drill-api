package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type limiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func getEnvInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}

func AuthRateLimiter() gin.HandlerFunc {
	rpm := getEnvInt("RATE_LIMIT_AUTH_RPM", 10)
	burst := getEnvInt("RATE_LIMIT_AUTH_BURST", 10)

	limit := rate.Limit(float64(rpm) / 60.0)

	return newRateLimiter(limit, burst)
}

func GenQuestionsRateLimiter() gin.HandlerFunc {
	rpm := getEnvInt("RATE_LIMIT_GEN_RPM", 3)
	burst := getEnvInt("RATE_LIMIT_GEN_BURST", 3)

	limit := rate.Limit(float64(rpm) / 60.0)

	return newRateLimiter(limit, burst)
}

func newRateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]*limiterClient)

	go func() {
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 1*time.Hour {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := clients[ip]
		if !exists {
			limiter := rate.NewLimiter(r, b)
			v = &limiterClient{limiter: limiter, lastSeen: time.Now()}
			clients[ip] = v
		} else {
			v.lastSeen = time.Now()
		}
		mu.Unlock()

		if !v.limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}
