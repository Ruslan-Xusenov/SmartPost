package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type uploadMediaResponse struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
}

func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request) {
	telegramIDStr := getTelegramID(r)
	if telegramIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chatID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Limit upload size to 50MB (Telegram's limit for bots)
	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mediaType := r.FormValue("media_type")
	if mediaType == "" {
		// default to photo
		mediaType = "photo"
	}

	var msg *models.Message
	var sendErr error

	switch mediaType {
	case "photo":
		msg, sendErr = s.botAPI.SendPhoto(r.Context(), &bot.SendPhotoParams{
			ChatID: chatID,
			Photo: &models.InputFileUpload{
				Filename: header.Filename,
				Data:     file,
			},
			Caption: "Bu rasmni siz post yaratish uchun yukladingiz (Qoralama).",
		})
	case "video":
		msg, sendErr = s.botAPI.SendVideo(r.Context(), &bot.SendVideoParams{
			ChatID: chatID,
			Video: &models.InputFileUpload{
				Filename: header.Filename,
				Data:     file,
			},
			Caption: "Bu videoni siz post yaratish uchun yukladingiz (Qoralama).",
		})
	case "video_note":
		msg, sendErr = s.botAPI.SendVideoNote(r.Context(), &bot.SendVideoNoteParams{
			ChatID: chatID,
			VideoNote: &models.InputFileUpload{
				Filename: header.Filename,
				Data:     file,
			},
		})
	default:
		// Fallback to document
		msg, sendErr = s.botAPI.SendDocument(r.Context(), &bot.SendDocumentParams{
			ChatID: chatID,
			Document: &models.InputFileUpload{
				Filename: header.Filename,
				Data:     file,
			},
			Caption: "Bu faylni siz post yaratish uchun yukladingiz (Qoralama).",
		})
	}

	if sendErr != nil {
		http.Error(w, fmt.Sprintf("Failed to send media to Telegram: %v", sendErr), http.StatusInternalServerError)
		return
	}

	var fileID string
	if msg.Photo != nil && len(msg.Photo) > 0 {
		// Get the largest photo size
		fileID = msg.Photo[len(msg.Photo)-1].FileID
	} else if msg.Video != nil {
		fileID = msg.Video.FileID
	} else if msg.VideoNote != nil {
		fileID = msg.VideoNote.FileID
	} else if msg.Document != nil {
		fileID = msg.Document.FileID
	}

	if fileID == "" {
		http.Error(w, "Failed to retrieve file_id from Telegram", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uploadMediaResponse{
		FileID:   fileID,
		FileName: header.Filename,
	})
}
