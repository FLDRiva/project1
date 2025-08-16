package handlers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"sellmind-backend/config"
	"sellmind-backend/models"
)

func AuthHandler(c *gin.Context) {
	var req struct {
		InitData string `json:"initData"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		c.JSON(500, gin.H{"error": "Bot token not set"})
		return
	}

	if !isValidTelegramInitData(req.InitData, botToken) {
		c.JSON(401, gin.H{"error": "Invalid signature"})
		return
	}

	user, err := parseUserFromInitData(req.InitData)
	if err != nil {
		c.JSON(400, gin.H{"error": "Can't parse user data"})
		return
	}

	err = upsertUser(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save user"})
		return
	}

	c.JSON(200, gin.H{
		"message": "authorized",
		"user":    user,
	})
}

func isValidTelegramInitData(initData, botToken string) bool {
	pairs := strings.Split(initData, "&")
	keys := []string{}
	params := make(map[string]string)

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key, _ := url.QueryUnescape(kv[0])
			value, _ := url.QueryUnescape(kv[1])
			params[key] = value
			if key != "hash" {
				keys = append(keys, kv[0]+"="+kv[1])
			}
		}
	}

	sort.Strings(keys)
	dataCheckString := strings.Join(keys, "\n")

	h := sha256.New()
	h.Write([]byte(botToken))
	secretKey := h.Sum(nil)

	h = sha256.New()
	h.Write(secretKey)
	h.Write([]byte("\n"))
	h.Write([]byte(dataCheckString))
	digest := h.Sum(nil)

	computedHash := fmt.Sprintf("%x", digest)
	return computedHash == params["hash"] && params["hash"] != ""
}

func parseUserFromInitData(initData string) (*models.User, error) {
	for _, pair := range strings.Split(initData, "&") {
		if strings.HasPrefix(pair, "user=") {
			raw := strings.TrimPrefix(pair, "user=")
			decoded, _ := url.QueryUnescape(raw)

			var user models.User
			user.TelegramID = extractInt(decoded, `id%22%3A(%d+)`)
			user.FirstName = extractString(decoded, `first_name%22%3A%22([^%]+)`)
			user.LastName = extractString(decoded, `last_name%22%3A%22([^%]+)`)
			user.Username = extractString(decoded, `username%22%3A%22([^%]+)`)
			user.LanguageCode = extractString(decoded, `language_code%22%3A%22([^%]+)`)

			return &user, nil
		}
	}
	return nil, fmt.Errorf("no user data")
}

func extractInt(data, pattern string) int64 {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(data)
	if len(matches) > 1 {
		var val int64
		fmt.Sscanf(matches[1], "%d", &val)
		return val
	}
	return 0
}

func extractString(data, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(data)
	if len(matches) > 1 {
		decoded, _ := url.QueryUnescape(matches[1])
		return strings.Split(decoded, "%")[0]
	}
	return ""
}

func upsertUser(u *models.User) error {
	query := `
		INSERT INTO users (telegram_id, first_name, last_name, username, language_code)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username,
			language_code = EXCLUDED.language_code,
			updated_at = NOW()
	`

	_, err := config.DB.Exec(
		context.Background(),
		query,
		u.TelegramID, u.FirstName, u.LastName, u.Username, u.LanguageCode,
	)
	return err
}