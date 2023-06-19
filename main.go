package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	// tgClient	"tg-bot/clients/telegram"
	// "tg-bot/storage/files"

	tgClient "tg-bot/clients/telegram"
	event_consumer "tg-bot/consumer/event-consumer"
	tgEvents "tg-bot/events/telegram"
	"tg-bot/storage/sqlite"
)

const (
	tgBotHost      = "api.telegram.org"
	storagePath    = "files_storage"
	storageSqlPath = "data/sqlite/storage.db"
	batchSize      = 100
)

func main() {
	token := mustToken()

	newTgClient := tgClient.New(tgBotHost, token)
	// storage := files.New(storagePath) // file storage
	storage, err := sqlite.New(storageSqlPath)
	if err != nil {
		log.Fatal("can't connect to storage: ", err)
	}
	// context.TODO - это аналог Background, он используется там, где мы пока ещё не знаем какой контекст будем юзать
	// Например, в будущем здесь может быть context.WithTimeout или что-то другое
	if err := storage.Init(context.TODO()); err != nil {
		log.Fatal("can't init storage", err)
	}

	eventsProcessor := tgEvents.New(newTgClient, storage)

	log.Print("service started")
	log.Print("eventsProcessor", eventsProcessor)

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}

// mustToken - Приставка must указывает на то, что функция не возвращает ошибку и параметр является обязательным
func mustToken() string {
	// bot -tg-bot-token 'my_token'
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access Telegram bot",
	)

	flag.Parse()

	fmt.Println("token", token)

	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token
}
