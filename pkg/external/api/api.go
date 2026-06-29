package api

import (
	"context"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	bot      *tgbotapi.BotAPI
	rdb      *redis.Client
	mu       sync.RWMutex
	logs     *log.Logger
	ctx      context.Context
}

func NewService(token string, rdb *redis.Client, ctx context.Context, logs *log.Logger) (*Service, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Service{
		bot:      bot,
		rdb:      rdb,
		logs:     logs,
		ctx:      ctx,
	}, nil
}

func (s *Service) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)

	s.logs.Printf("Bot started, waiting for updates...\n")

	for update := range updates {
		if update.Message != nil {
			s.logs.Printf("Received message from %s: %s\n", update.Message.From.UserName, update.Message.Text)
			s.handleMessage(update.Message)
		}
	}

	return nil
}
