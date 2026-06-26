package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
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
	// Parse the query string
	params := parseQueryString(initData)

	hash, ok := params["hash"]
	if !ok {
		return false
	}

	// Remove hash from params and sort
	delete(params, "hash")
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build data-check-string
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
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

func parseQueryString(qs string) map[string]string {
	params := make(map[string]string)
	for _, pair := range strings.Split(qs, "&") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}
	return params
}

func extractUserID(initData string) string {
	// Simple extraction - in production, decode the JSON user object
	params := parseQueryString(initData)
	if user, ok := params["user"]; ok {
		// user is URL-encoded JSON, extract id field
		// For simplicity, find "id": in the string
		idx := strings.Index(user, "%22id%22%3A")
		if idx != -1 {
			rest := user[idx+len("%22id%22%3A"):]
			end := strings.IndexAny(rest, "%,}")
			if end != -1 {
				return rest[:end]
			}
		}
		// Try non-encoded format
		idx = strings.Index(user, `"id":`)
		if idx != -1 {
			rest := user[idx+5:]
			end := strings.IndexAny(rest, ",}")
			if end != -1 {
				return strings.TrimSpace(rest[:end])
			}
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
