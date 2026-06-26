package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type contextKey string

const userTelegramIDKey contextKey = "telegram_id"

// twaAuthMiddleware validates Telegram Web App initData.
func (s *Server) twaAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			http.Error(w, "Missing init data", http.StatusUnauthorized)
			return
		}

		// Validate initData using HMAC-SHA256
		if !s.validateInitData(initData) {
			http.Error(w, "Invalid init data", http.StatusUnauthorized)
			return
		}

		// Extract user ID from initData
		telegramID := extractUserID(initData)
		if telegramID == "" {
			http.Error(w, "Invalid user data", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userTelegramIDKey, telegramID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateInitData checks the HMAC-SHA256 signature of Telegram initData.
func (s *Server) validateInitData(initData string) bool {
	// Parse the query string (this handles URL decoding)
	params, err := url.ParseQuery(initData)
	if err != nil {
		return false
	}

	hash := params.Get("hash")
	if hash == "" {
		return false
	}

	// Remove hash from params and sort keys
	params.Del("hash")
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build data-check-string
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params.Get(k))
	}
	dataCheckString := strings.Join(parts, "\n")

	// Create secret key: HMAC-SHA256("WebAppData", bot_token)
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(s.cfg.BotToken))

	// Verify: HMAC-SHA256(secret_key, data_check_string) == hash
	expectedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))
	return expectedHash == hash
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func extractUserID(initData string) string {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return ""
	}
	
	userStr := params.Get("user")
	if userStr == "" {
		return ""
	}
	
	// Parse JSON
	var user map[string]interface{}
	if err := json.Unmarshal([]byte(userStr), &user); err == nil {
		if id, ok := user["id"].(float64); ok {
			return fmt.Sprintf("%.0f", id)
		}
	}
	return ""
}

// getTelegramID extracts the authenticated user's Telegram ID from context.
func getTelegramID(r *http.Request) string {
	if v, ok := r.Context().Value(userTelegramIDKey).(string); ok {
		return v
	}
	return ""
}
