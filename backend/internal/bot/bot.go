package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/smartpost/backend/internal/config"
	"github.com/smartpost/backend/internal/scheduler"
)

type Bot struct {
	api    *bot.Bot
	db     *pgxpool.Pool
	rdb       *redis.Client
	cfg       *config.Config
	fsm       *FSM
	publisher *scheduler.Publisher
}

func New(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client, publisher *scheduler.Publisher) (*Bot, error) {
	b := &Bot{
		db:        db,
		rdb:       rdb,
		cfg:       cfg,
		fsm:       NewFSM(rdb),
		publisher: publisher,
	}
	opts := []bot.Option{
		bot.WithDefaultHandler(b.defaultHandler),
		bot.WithCallbackQueryDataHandler("", bot.MatchTypePrefix, b.callbackHandler),
	}

	api, err := bot.New(cfg.BotToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	b.api = api
	b.registerHandlers()

	return b, nil
}

func (b *Bot) registerHandlers() {
	b.api.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.handleStart)
	b.api.RegisterHandler(bot.HandlerTypeMessageText, "/newpost", bot.MatchTypeExact, b.handleNewPost)
	b.api.RegisterHandler(bot.HandlerTypeMessageText, "/mychannels", bot.MatchTypeExact, b.handleMyChannels)
	b.api.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, b.handleHelp)
	b.api.RegisterHandler(bot.HandlerTypeMessageText, "/cancel", bot.MatchTypeExact, b.handleCancel)
}

func (b *Bot) SetWebhook(ctx context.Context) error {
	if b.cfg.WebhookURL == "" {
		// Local development: delete webhook to enable long polling
		_, err := b.api.DeleteWebhook(ctx, &bot.DeleteWebhookParams{})
		if err != nil {
			return fmt.Errorf("failed to delete webhook for polling: %w", err)
		}
		log.Println("✅ Webhook deleted (Local Polling mode enabled)")
		return nil
	}

	ok, err := b.api.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: b.cfg.WebhookURL,
	})
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}
	if !ok {
		return fmt.Errorf("webhook was not set")
	}
	log.Printf("✅ Webhook set: %s", b.cfg.WebhookURL)
	return nil
}

func (b *Bot) WebhookHandler() func(ctx context.Context, bBot *bot.Bot, update *models.Update) {
	return func(ctx context.Context, bBot *bot.Bot, update *models.Update) {
		if update.MyChatMember != nil {
			b.handleMyChatMember(ctx, update.MyChatMember)
			return
		}
	}
}

func (b *Bot) Start(ctx context.Context) {
	if b.cfg.WebhookURL == "" {
		b.api.Start(ctx)
	}
	// In webhook mode, the API server handles incoming requests.
}

func (b *Bot) HandleWebhookUpdate(ctx context.Context, update *models.Update) {
	b.api.ProcessUpdate(ctx, update)
}

func (b *Bot) API() *bot.Bot {
	return b.api
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string) {
	_, err := b.api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("❌ Failed to send message to %d: %v", chatID, err)
	}
}

func (b *Bot) sendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard *models.InlineKeyboardMarkup) {
	_, err := b.api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		log.Printf("❌ Failed to send message with keyboard to %d: %v", chatID, err)
	}
}