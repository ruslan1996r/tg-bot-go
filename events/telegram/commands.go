package tg_events

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"

	tgClient "tg-bot/clients/telegram"
	"tg-bot/lib/e"
	"tg-bot/storage"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start "
)

// doCmd - обрабатывает команды вроде help
func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text) // Убрать пробелы в начале и конце

	log.Printf("got new command '%s' from '%s'", text, username)

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		return p.tg.SendMessages(chatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	sendMsg := NewMessageSender(chatID, p.tg) // Функция используется больше для примера

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	// TODO: По нормальной практике Контекст всегда надо передавать сверху savePage(ctx, ...)
	// Чтобы сверху можно было остановить все вложенные функции
	isExists, err := p.storage.IsExists(context.Background(), page)
	if isExists {
		// return p.tg.SendMessages(chatID, msgAlreadyExists)
		return sendMsg(msgAlreadyExists) // В этом случае не нужно постоянно передавать chatID
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	// if err := sendMsg(msgSaved); err != nil { // Аналогичный тип записи
	if err := p.tg.SendMessages(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	// Достань случайную ссылку
	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	// Если ссылка не была найдена, верни сообщение в ТГ, что такой ссылки не
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessages(chatID, msgNoSavedPages)
	}

	// Если ссылка была найдена, отправь её
	if err := p.tg.SendMessages(chatID, page.URL); err != nil {
		return err
	}

	// Если ссылка была найдена и отправлена, её нужно удалить
	return p.storage.Remove(context.Background(), page)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessages(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessages(chatID, msgHello)
}

// NewMessageSender - такой подход позволяет переиспользовать краткую запись функции
// Пример работает на замыкании. То есть мы 1 раз передаём параметры, инитим функцию и переюзаем её
func NewMessageSender(chatID int, tg *tgClient.Client) func(string) error {
	return func(msg string) error {
		return tg.SendMessages(chatID, msg)
	}
}

func isAddCmd(text string) bool {
	return isURL(text)
}

// isURL - за URL считается та строка, которая не возвращает ошибку и имеет в себе Host (не покрывает все кейсы)
func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
