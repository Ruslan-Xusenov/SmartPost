package api

import (
	"encoding/json"
	"net/http"
)

// listChannels returns all channels owned by the authenticated user.
func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	telegramID := getTelegramID(r)

	rows, err := s.db.Query(r.Context(),
		`SELECT c.id, c.chat_id, c.title, c.username, c.is_active
		 FROM channels c
		 JOIN users u ON u.id = c.owner_id
		 WHERE u.telegram_id = $1 AND c.is_active = true
		 ORDER BY c.title`, telegramID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type channelResp struct {
		ID       int64  `json:"id"`
		ChatID   int64  `json:"chat_id"`
		Title    string `json:"title"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	}

	var channels []channelResp
	for rows.Next() {
		var ch channelResp
		if err := rows.Scan(&ch.ID, &ch.ChatID, &ch.Title, &ch.Username, &ch.IsActive); err != nil {
			continue
		}
		channels = append(channels, ch)
	}

	if channels == nil {
		channels = []channelResp{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}
