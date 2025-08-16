package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type GenerateResponseRequest struct {
	ReviewText string `json:"review_text"`
	Rating     int    `json:"rating"`
}

type HFResponse []map[string]string

func GenerateResponseHandler(c *gin.Context) {
	var req GenerateResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	hfToken := os.Getenv("HF_API_TOKEN")
	if hfToken == "" {
		c.JSON(500, gin.H{"error": "HF_API_TOKEN not set"})
		return
	}

	// Простой промт
	prompt := fmt.Sprintf("Напиши вежливый ответ продавца на отзыв: %s", req.ReviewText)

	payload := map[string]any{
		"inputs": prompt,
		"parameters": map[string]any{
			"max_new_tokens": 100,
			"temperature":    0.7,
			"top_p":          0.9,
			"do_sample":      true,
		},
	}

	jsonData, _ := json.Marshal(payload)

	// ✅ Рабочий URL
	url := "https://api-inference.huggingface.co/models/google/flan-t5-base"

	reqHTTP, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	reqHTTP.Header.Set("Authorization", "Bearer "+hfToken)
	reqHTTP.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		c.JSON(500, gin.H{"error": "Request failed"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		c.JSON(500, gin.H{
			"error":    "Model error",
			"status":   resp.Status,
			"response": string(body),
		})
		return
	}

	var result HFResponse
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(500, gin.H{"error": "Parse error"})
		return
	}

	if len(result) == 0 {
		c.JSON(500, gin.H{"error": "Empty response"})
		return
	}

	full := result[0]["generated_text"]
	response := strings.TrimPrefix(full, prompt)
	response = strings.TrimSpace(response)

	// Обрезаем до первого предложения
	if idx := strings.IndexAny(response, ".!?"); idx != -1 {
		response = response[:idx+1]
	}

	if response == "" {
		response = "Спасибо за ваш отзыв!"
	}

	c.JSON(200, gin.H{"response": response})
}