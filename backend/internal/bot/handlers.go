package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	mdl "github.com/smartpost/backend/internal/models"
)

// handleStart registers a new user or greets an existing one.
func (b *Bot) handleStart(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	user := update.Message.From
	_, err := b.db.Exec(ctx,
		`INSERT INTO users (telegram_id, username, first_name, timezone)
		 VALUES ($1, $2, $3, 'Asia/Tashkent')
		 ON CONFLICT (telegram_id) DO UPDATE SET username = $2, first_name = $3, updated_at = NOW()`,
		user.ID, user.Username, user.FirstName,
	)
	if err != nil {
		log.Printf("❌ Failed to upsert user %d: %v", user.ID, err)
	}

	text := fmt.Sprintf(
		"👋 Salom, <b>%s</b>!\n\n"+
			"🤖 <b>SmartPost</b> — Telegram kanallaringizni boshqarish tizimi.\n\n"+
			"📌 <b>Buyruqlar:</b>\n"+
			"/newpost — Yangi post yaratish\n"+
			"/mychannels — Kanallarim ro'yxati\n"+
			"/help — Yordam\n\n"+
			"⚡ Boshlash uchun botni kanalingizga <b>admin</b> sifatida qo'shing!",
		user.FirstName,
	)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "📝 Post yaratish", WebApp: &models.WebAppInfo{URL: b.cfg.TWAURL}},
			},
			{
				{Text: "📢 Kanallarim", CallbackData: "channels_list"},
			},
		},
	}
	b.sendMessageWithKeyboard(ctx, update.Message.Chat.ID, text, keyboard)
}

// handleNewPost starts the post creation FSM flow.
func (b *Bot) handleNewPost(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Get user's channels
	rows, err := b.db.Query(ctx,
		`SELECT c.id, c.chat_id, c.title FROM channels c
		 JOIN users u ON u.id = c.owner_id
		 WHERE u.telegram_id = $1 AND c.is_active = true`, userID)
	if err != nil {
		b.sendMessage(ctx, chatID, "❌ Xatolik yuz berdi.")
		return
	}
	defer rows.Close()

	var buttons [][]models.InlineKeyboardButton
	hasChannels := false
	for rows.Next() {
		var ch mdl.Channel
		if err := rows.Scan(&ch.ID, &ch.ChatID, &ch.Title); err != nil {
			continue
		}
		hasChannels = true
		buttons = append(buttons, []models.InlineKeyboardButton{
			{Text: "📢 " + ch.Title, CallbackData: fmt.Sprintf("sel_ch:%d", ch.ID)},
		})
	}

	if !hasChannels {
		b.sendMessage(ctx, chatID, "⚠️ Sizda hali kanal yo'q!\n\nBotni kanalingizga <b>admin</b> sifatida qo'shing.")
		return
	}

	// Set FSM state to select channel
	if err := b.fsm.SetState(ctx, userID, StateSelectChannel); err != nil {
		log.Printf("❌ FSM error: %v", err)
	}

	b.sendMessageWithKeyboard(ctx, chatID, "📢 Qaysi kanalga post yubormoqchisiz?", &models.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	})
}

// handleMyChannels lists the user's linked channels.
func (b *Bot) handleMyChannels(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	rows, err := b.db.Query(ctx,
		`SELECT c.title, c.username, c.is_active FROM channels c
		 JOIN users u ON u.id = c.owner_id
		 WHERE u.telegram_id = $1`, userID)
	if err != nil {
		b.sendMessage(ctx, chatID, "❌ Xatolik yuz berdi.")
		return
	}
	defer rows.Close()

	var text strings.Builder
	text.WriteString("📢 <b>Sizning kanallaringiz:</b>\n\n")
	count := 0
	for rows.Next() {
		var title, username string
		var isActive bool
		if err := rows.Scan(&title, &username, &isActive); err != nil {
			continue
		}
		count++
		status := "✅"
		if !isActive {
			status = "❌"
		}
		text.WriteString(fmt.Sprintf("%s <b>%s</b>", status, title))
		if username != "" {
			text.WriteString(fmt.Sprintf(" (@%s)", username))
		}
		text.WriteString("\n")
	}

	if count == 0 {
		text.Reset()
		text.WriteString("⚠️ Sizda hali kanal yo'q.\n\nBotni kanalingizga <b>admin</b> sifatida qo'shing.")
	}
	b.sendMessage(ctx, chatID, text.String())
}

// handleHelp sends help information.
func (b *Bot) handleHelp(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	text := "ℹ️ <b>SmartPost - Yordam</b>\n\n" +
		"<b>Buyruqlar:</b>\n" +
		"/start — Botni boshlash\n" +
		"/newpost — Yangi post yaratish\n" +
		"/mychannels — Kanallar ro'yxati\n" +
		"/cancel — Jarayonni bekor qilish\n\n" +
		"<b>Post yaratish:</b>\n" +
		"1. /newpost buyrug'ini yuboring\n" +
		"2. Kanalni tanlang\n" +
		"3. Media yuboring (rasm, video, video-note)\n" +
		"4. Matn/caption yozing\n" +
		"5. Tugmalar qo'shing (ixtiyoriy)\n" +
		"   Format: <code>Tugma matni - URL - rang</code>\n" +
		"6. Vaqt tanlang yoki darhol yuboring\n\n" +
		"<b>Ranglar:</b> green, red, blue, yellow, orange, purple"
	b.sendMessage(ctx, update.Message.Chat.ID, text)
}

// handleCancel cancels the current FSM operation.
func (b *Bot) handleCancel(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	_ = b.fsm.Clear(ctx, update.Message.From.ID)
	b.sendMessage(ctx, update.Message.Chat.ID, "❌ Jarayon bekor qilindi.")
}

// handleMyChatMember processes bot being added/removed from channels.
func (b *Bot) handleMyChatMember(ctx context.Context, member *models.ChatMemberUpdated) {
	chat := member.Chat
	from := member.From
	newStatus := member.NewChatMember

	// Only process channel events
	if chat.Type != "channel" {
		return
	}

	// Check if bot was made admin
	if newStatus.Type == models.ChatMemberTypeAdministrator || newStatus.Type == models.ChatMemberTypeOwner {
		// Get or create user
		var userDBID int64
		err := b.db.QueryRow(ctx,
			`SELECT id FROM users WHERE telegram_id = $1`, from.ID,
		).Scan(&userDBID)
		if err != nil {
			log.Printf("⚠️ User %d not found, cannot link channel %d", from.ID, chat.ID)
			return
		}

		_, err = b.db.Exec(ctx,
			`INSERT INTO channels (owner_id, chat_id, title, username, is_active)
			 VALUES ($1, $2, $3, $4, true)
			 ON CONFLICT (chat_id) DO UPDATE SET
			   owner_id = $1, title = $3, username = $4, is_active = true`,
			userDBID, chat.ID, chat.Title, chat.Username,
		)
		if err != nil {
			log.Printf("❌ Failed to link channel: %v", err)
			return
		}
		log.Printf("✅ Channel linked: %s (%d) by user %d", chat.Title, chat.ID, from.ID)

		b.sendMessage(ctx, from.ID,
			fmt.Sprintf("✅ Kanal ulandi: <b>%s</b>\n\nEndi /newpost orqali post yaratishingiz mumkin!", chat.Title))
	}

	// Check if bot was removed
	if newStatus.Type == models.ChatMemberTypeLeft || newStatus.Type == models.ChatMemberTypeBanned {
		_, err := b.db.Exec(ctx,
			`UPDATE channels SET is_active = false WHERE chat_id = $1`, chat.ID)
		if err != nil {
			log.Printf("❌ Failed to deactivate channel: %v", err)
		}
		log.Printf("⚠️ Bot removed from channel: %s (%d)", chat.Title, chat.ID)
	}
}

// callbackHandler processes inline keyboard button presses.
func (b *Bot) callbackHandler(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}
	data := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	// Answer callback to remove loading state
	tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	switch {
	case strings.HasPrefix(data, "sel_ch:"):
		b.handleChannelSelected(ctx, userID, chatID, data)
	case data == "skip_caption":
		b.handleSkipCaption(ctx, userID, chatID)
	case data == "skip_buttons":
		b.handleSkipButtons(ctx, userID, chatID)
	case data == "send_now":
		b.handleSendNow(ctx, userID, chatID)
	case data == "confirm_send":
		b.handleConfirmSend(ctx, userID, chatID)
	case data == "cancel_post":
		_ = b.fsm.Clear(ctx, userID)
		b.sendMessage(ctx, chatID, "❌ Post bekor qilindi.")
	case data == "channels_list":
		b.handleChannelsListCallback(ctx, userID, chatID)
	}
}

// handleChannelSelected processes channel selection in FSM.
func (b *Bot) handleChannelSelected(ctx context.Context, userID, chatID int64, data string) {
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return
	}
	channelDBID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	fsmData, err := b.fsm.Get(ctx, userID)
	if err != nil || fsmData.State != StateSelectChannel {
		return
	}

	fsmData.ChannelID = channelDBID
	fsmData.State = StateWaitMedia
	if err := b.fsm.Set(ctx, userID, fsmData); err != nil {
		log.Printf("❌ FSM error: %v", err)
		return
	}

	b.sendMessage(ctx, chatID, "📎 Endi media yuboring:\n\n"+
		"• 📷 Rasm\n• 🎥 Video\n• 🔵 Dumaloq video (video-note)\n• 📝 Yoki faqat matn yuboring")
}

func (b *Bot) handleSkipCaption(ctx context.Context, userID, chatID int64) {
	fsmData, err := b.fsm.Get(ctx, userID)
	if err != nil {
		return
	}
	fsmData.State = StateWaitButtons
	_ = b.fsm.Set(ctx, userID, fsmData)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "⏭ O'tkazish", CallbackData: "skip_buttons"}},
		},
	}
	b.sendMessageWithKeyboard(ctx, chatID,
		"🔘 Inline tugmalar qo'shing.\n\nFormat (har bir satrda bitta):\n"+
			"<code>Tugma matni - URL - rang</code>\n\n"+
			"Ranglar: green, red, blue, yellow, orange, purple\n"+
			"Misol: <code>Saytga o'tish - https://example.com - green</code>",
		keyboard,
	)
}

func (b *Bot) handleSkipButtons(ctx context.Context, userID, chatID int64) {
	fsmData, err := b.fsm.Get(ctx, userID)
	if err != nil {
		return
	}
	fsmData.State = StateWaitSchedule
	_ = b.fsm.Set(ctx, userID, fsmData)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "🚀 Hozir yuborish", CallbackData: "send_now"}},
		},
	}
	b.sendMessageWithKeyboard(ctx, chatID,
		"⏰ Qachon yuborilsin?\n\nVaqtni kiriting (format: <code>2025-01-15 14:30</code>)\nyoki hozir yuboring:",
		keyboard,
	)
}

func (b *Bot) handleSendNow(ctx context.Context, userID, chatID int64) {
	fsmData, err := b.fsm.Get(ctx, userID)
	if err != nil {
		return
	}
	fsmData.Schedule = "now"
	fsmData.State = StateConfirm
	_ = b.fsm.Set(ctx, userID, fsmData)
	b.showConfirmation(ctx, userID, chatID, fsmData)
}

func (b *Bot) handleConfirmSend(ctx context.Context, userID, chatID int64) {
	fsmData, err := b.fsm.Get(ctx, userID)
	if err != nil || fsmData.State != StateConfirm {
		return
	}

	// Get user DB ID
	var userDBID int64
	err = b.db.QueryRow(ctx, `SELECT id FROM users WHERE telegram_id = $1`, userID).Scan(&userDBID)
	if err != nil {
		b.sendMessage(ctx, chatID, "❌ Foydalanuvchi topilmadi.")
		return
	}

	// Create post in database
	var postID int64
	err = b.db.QueryRow(ctx,
		`INSERT INTO posts (channel_id, user_id, media_type, file_id, caption, status)
		 VALUES ($1, $2, $3, $4, $5, 'draft') RETURNING id`,
		fsmData.ChannelID, userDBID, fsmData.MediaType, fsmData.FileID, fsmData.Caption,
	).Scan(&postID)
	if err != nil {
		b.sendMessage(ctx, chatID, "❌ Post saqlashda xatolik.")
		log.Printf("❌ Post insert error: %v", err)
		return
	}

	// Parse and save buttons
	if fsmData.Buttons != "" {
		buttons := mdl.ParseButtons(fsmData.Buttons)
		for _, btn := range buttons {
			_, err := b.db.Exec(ctx,
				`INSERT INTO buttons (post_id, text, url, color_code, row_index, col_index)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				postID, btn.Text, btn.URL, btn.ColorCode, btn.RowIndex, btn.ColIndex,
			)
			if err != nil {
				log.Printf("❌ Button insert error: %v", err)
			}
		}
	}

	// Schedule or send immediately
	if fsmData.Schedule == "now" {
		_, _ = b.db.Exec(ctx, `UPDATE posts SET status = 'scheduled' WHERE id = $1`, postID)
		if err := b.publisher.EnqueuePostSend(postID); err != nil {
			log.Printf("❌ Failed to enqueue post: %v", err)
			b.sendMessage(ctx, chatID, "❌ Postni yuborishda xatolik yuz berdi.")
		} else {
			b.sendMessage(ctx, chatID, "✅ Post yuborilmoqda...")
		}
	} else {
		_, _ = b.db.Exec(ctx,
			`UPDATE posts SET status = 'scheduled', scheduled_at = $1 WHERE id = $2`,
			fsmData.Schedule, postID,
		)
		
		t, err := time.Parse("2006-01-02 15:04", fsmData.Schedule)
		if err == nil {
			if err := b.publisher.SchedulePostSend(postID, t); err != nil {
				log.Printf("❌ Failed to schedule post: %v", err)
			}
		}
		
		b.sendMessage(ctx, chatID, fmt.Sprintf("✅ Post rejalashtirildi: <b>%s</b>", fsmData.Schedule))
	}

	_ = b.fsm.Clear(ctx, userID)
}

func (b *Bot) showConfirmation(ctx context.Context, userID, chatID int64, fsmData *FSMData) {
	var text strings.Builder
	text.WriteString("📋 <b>Post tasdiqlash:</b>\n\n")
	text.WriteString(fmt.Sprintf("📎 Media: <b>%s</b>\n", fsmData.MediaType))
	if fsmData.Caption != "" {
		text.WriteString(fmt.Sprintf("📝 Matn: %s\n", fsmData.Caption))
	}
	if fsmData.Buttons != "" {
		text.WriteString(fmt.Sprintf("🔘 Tugmalar: %d ta\n", len(strings.Split(strings.TrimSpace(fsmData.Buttons), "\n"))))
	}
	if fsmData.Schedule == "now" {
		text.WriteString("⏰ Yuborish: <b>Hozir</b>\n")
	} else {
		text.WriteString(fmt.Sprintf("⏰ Yuborish: <b>%s</b>\n", fsmData.Schedule))
	}

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Tasdiqlash", CallbackData: "confirm_send"},
				{Text: "❌ Bekor qilish", CallbackData: "cancel_post"},
			},
		},
	}
	b.sendMessageWithKeyboard(ctx, chatID, text.String(), keyboard)
}

func (b *Bot) handleChannelsListCallback(ctx context.Context, userID, chatID int64) {
	rows, err := b.db.Query(ctx,
		`SELECT c.title, c.username, c.is_active FROM channels c
		 JOIN users u ON u.id = c.owner_id
		 WHERE u.telegram_id = $1`, userID)
	if err != nil {
		b.sendMessage(ctx, chatID, "❌ Xatolik.")
		return
	}
	defer rows.Close()

	var text strings.Builder
	text.WriteString("📢 <b>Kanallaringiz:</b>\n\n")
	count := 0
	for rows.Next() {
		var title, username string
		var isActive bool
		if err := rows.Scan(&title, &username, &isActive); err != nil {
			continue
		}
		count++
		status := "✅"
		if !isActive {
			status = "❌"
		}
		text.WriteString(fmt.Sprintf("%s %s", status, title))
		if username != "" {
			text.WriteString(fmt.Sprintf(" (@%s)", username))
		}
		text.WriteString("\n")
	}
	if count == 0 {
		text.Reset()
		text.WriteString("⚠️ Hali kanal yo'q. Botni kanalga admin qiling!")
	}
	b.sendMessage(ctx, chatID, text.String())
}

// defaultHandler processes all non-command messages (media, text during FSM).
func (b *Bot) defaultHandler(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	// Process channel status changes (when bot is added/removed as admin)
	if update.MyChatMember != nil {
		b.handleMyChatMember(ctx, update.MyChatMember)
		return
	}

	if update.Message == nil {
		return
	}
	msg := update.Message
	userID := msg.From.ID
	chatID := msg.Chat.ID

	fsmData, err := b.fsm.Get(ctx, userID)
	if err != nil || fsmData.State == StateIdle {
		return
	}

	switch fsmData.State {
	case StateWaitMedia:
		b.handleMediaInput(ctx, userID, chatID, msg, fsmData)
	case StateWaitCaption:
		b.handleCaptionInput(ctx, userID, chatID, msg, fsmData)
	case StateWaitButtons:
		b.handleButtonsInput(ctx, userID, chatID, msg, fsmData)
	case StateWaitSchedule:
		b.handleScheduleInput(ctx, userID, chatID, msg, fsmData)
	}
}

func (b *Bot) handleMediaInput(ctx context.Context, userID, chatID int64, msg *models.Message, fsmData *FSMData) {
	switch {
	case msg.Photo != nil && len(msg.Photo) > 0:
		// Get the largest photo
		photo := msg.Photo[len(msg.Photo)-1]
		fsmData.MediaType = string(mdl.MediaTypePhoto)
		fsmData.FileID = photo.FileID
	case msg.Video != nil:
		fsmData.MediaType = string(mdl.MediaTypeVideo)
		fsmData.FileID = msg.Video.FileID
	case msg.VideoNote != nil:
		fsmData.MediaType = string(mdl.MediaTypeVideoNote)
		fsmData.FileID = msg.VideoNote.FileID
	case msg.Text != "":
		fsmData.MediaType = string(mdl.MediaTypeText)
		fsmData.Caption = msg.Text
		fsmData.State = StateWaitButtons
		_ = b.fsm.Set(ctx, userID, fsmData)
		keyboard := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "⏭ O'tkazish", CallbackData: "skip_buttons"}},
			},
		}
		b.sendMessageWithKeyboard(ctx, chatID,
			"🔘 Inline tugmalar qo'shing.\n\nFormat:\n<code>Tugma matni - URL - rang</code>\n\nMisol:\n<code>Saytga o'tish - https://example.com - green</code>",
			keyboard)
		return
	default:
		b.sendMessage(ctx, chatID, "⚠️ Iltimos, rasm, video, dumaloq video yoki matn yuboring.")
		return
	}

	fsmData.State = StateWaitCaption
	_ = b.fsm.Set(ctx, userID, fsmData)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "⏭ O'tkazish", CallbackData: "skip_caption"}},
		},
	}
	b.sendMessageWithKeyboard(ctx, chatID, "📝 Endi matn/caption yozing yoki o'tkazing:", keyboard)
}

func (b *Bot) handleCaptionInput(ctx context.Context, userID, chatID int64, msg *models.Message, fsmData *FSMData) {
	if msg.Text == "" {
		b.sendMessage(ctx, chatID, "⚠️ Iltimos, matn yuboring.")
		return
	}
	fsmData.Caption = msg.Text
	fsmData.State = StateWaitButtons
	_ = b.fsm.Set(ctx, userID, fsmData)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "⏭ O'tkazish", CallbackData: "skip_buttons"}},
		},
	}
	b.sendMessageWithKeyboard(ctx, chatID,
		"🔘 Inline tugmalar qo'shing.\n\nFormat:\n<code>Tugma matni - URL - rang</code>",
		keyboard)
}

func (b *Bot) handleButtonsInput(ctx context.Context, userID, chatID int64, msg *models.Message, fsmData *FSMData) {
	if msg.Text == "" {
		b.sendMessage(ctx, chatID, "⚠️ Iltimos, tugmalar matnini yuboring.")
		return
	}
	fsmData.Buttons = msg.Text
	fsmData.State = StateWaitSchedule
	_ = b.fsm.Set(ctx, userID, fsmData)

	// Show parsed buttons preview
	buttons := mdl.ParseButtons(msg.Text)
	var preview strings.Builder
	preview.WriteString("✅ Tugmalar qabul qilindi:\n\n")
	for _, btn := range buttons {
		preview.WriteString(fmt.Sprintf("  %s → %s\n", btn.DisplayText(), btn.URL))
	}
	preview.WriteString("\n⏰ Qachon yuborilsin?")

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "🚀 Hozir yuborish", CallbackData: "send_now"}},
		},
	}
	b.sendMessageWithKeyboard(ctx, chatID, preview.String(), keyboard)
}

func (b *Bot) handleScheduleInput(ctx context.Context, userID, chatID int64, msg *models.Message, fsmData *FSMData) {
	if msg.Text == "" {
		b.sendMessage(ctx, chatID, "⚠️ Iltimos, vaqtni kiriting (format: 2025-01-15 14:30)")
		return
	}
	fsmData.Schedule = msg.Text
	fsmData.State = StateConfirm
	_ = b.fsm.Set(ctx, userID, fsmData)
	b.showConfirmation(ctx, userID, chatID, fsmData)
}
