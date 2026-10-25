package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

var recaptchaVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

func RecaptchaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appEnv := os.Getenv("APP_ENV")
		if appEnv == "local" || appEnv == "development" {
			c.Next()
			return
		}

		secret := os.Getenv("RECAPTCHA_SECRET")
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "reCAPTCHA configuration error: secret is missing",
			})
			return
		}

		token := c.GetHeader("X-Recaptcha-Token")
		if token == "" {
			token = c.Query("recaptcha_token")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "reCAPTCHA token is required",
			})
			return
		}

		resp, err := client.PostForm(recaptchaVerifyURL, url.Values{
			"secret":   {secret},
			"response": {token},
			"remoteip": {c.ClientIP()},
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to contact reCAPTCHA verification service",
			})
			return
		}
		defer resp.Body.Close()

		var result RecaptchaResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to decode reCAPTCHA response",
			})
			return
		}

		if !result.Success {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "reCAPTCHA verification failed",
				"errors":  result.ErrorCodes,
			})
			return
		}

		if minScoreStr := "0.5"; minScoreStr != "" {
			minScore, err := strconv.ParseFloat(minScoreStr, 64)
			if err == nil {
				if result.Score < minScore {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"success": false,
						"message": fmt.Sprintf("reCAPTCHA score %.2f is below the minimum threshold of %.2f", result.Score, minScore),
					})
					return
				}
			}
		}

		c.Next()
	}
}
