package event_consumer

import (
	"log"
	"time"

	"tg-bot/events"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int // Сколько событий обрабатывать за раз
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

// Start -  Содержит вечный цикл, который ждёт события и обрабатывает их
func (c Consumer) Start() error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)

		if err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())

			continue
		}

		// Если ивентов нет, засни на 1 секунду, а потом посмотри или есть новые ивенты
		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)

			continue
		}
	}
}

// type SingleEvent = events.Event

// handleEvents - проходит по пачке ивентов и обрабатывает их
// Если натыкается на ошибку в обработке ивента, просто пропускает её
// todo: add sync.WaitGroup{} для параллельной обработки
func (c Consumer) handleEvents(events []events.Event) error {
	for _, event := range events {
		log.Printf("got new event: %s", event.Text)

		// todo: как улучшить? все необработанные ивенты можно добавлять в какой-то механизм back-up (в какое-то хранилище) или retry
		if err := c.processor.Process(event); err != nil {
			log.Printf("can't handle event: %s", err.Error())

			continue
		}
	}

	return nil
}
