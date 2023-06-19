package tg_events

import (
	"errors"

	telegram "tg-bot/clients/telegram"
	"tg-bot/events"
	"tg-bot/lib/e"
	"tg-bot/storage"
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownErrType  = errors.New("unknown event type")
	ErrUnknownMetaType = errors.New("unknown meta type")
)

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	// Получить апдейты
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	// Если пустые апдейты, верни nil
	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	// Преобразовать все апдейты в тип Event
	for _, u := range updates {
		res = append(res, event(u))
	}

	// Обновить offset, чтобы в следующий раз получить следующую пачку сообщений
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

// Process - этот подход позволяет в будущем расширить функционал, если предстоит работать с другими типами ивентов
// Например, если будет что-то кроме Message
func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can't process message", ErrUnknownErrType)
	}
}

func (p *Processor) processMessage(event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	// TODO: почему-то Telegram API НЕ ВОЗВРАЩАЕТ chatID, поэтому я его просто захардкодил
	chatId := 211637098 // meta.ChatID

	if err := p.doCmd(event.Text, chatId, meta.Username); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

// meta - Утверждение типа event.Meta.(Meta). Проверяет или значение является нужным типом. Вернёт "bool"
func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta) // Type assertion
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}
	return upd.Message.Text
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}
	return events.Message
}
