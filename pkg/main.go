package main

import (
	"context"
	"log"
	"os"
	"telegramTestBot/pkg/external/api"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	_ "modernc.org/sqlite"
)

const (
	TokenEnv = "TOKEN"
)

func main() {
	logs := log.New(os.Stdout, "telegram-bot: ", log.LstdFlags)

	godotenv.Load("config.env")
	token := os.Getenv(TokenEnv)
	logs.Printf("Token acquired: %s\n", token)

	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}
	logs.Printf("Connected to Redis\n")

	service, err := api.NewService(token, rdb, ctx, logs)
	if err != nil {
		panic(err)
	}
	logs.Printf("Telegram bot service initialized\n")

	if err = service.Start(); err != nil {
		logs.Printf("Error starting telegram bot: %s\n", err)
	}
}
