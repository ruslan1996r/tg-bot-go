package tgClient

// UpdateResponse - То, как будет выглядеть успешный ответ с сервера
type UpdateResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

// Update - https://core.telegram.org/bots/api#getting-updates
// *IncomingMessage - указать ссылку на структуру (делает поле опциональным, если это объект)
type Update struct {
	ID      int              `json:"update_id"`
	Message *IncomingMessage `json:"message"`
}

type IncomingMessage struct {
	Text string `json:"text"`
	From From   `json:"from"`
	Chat Chat   `json:"chat"`
}

type From struct {
	Username string `json:"username"`
}

type Chat struct {
	ID int `json:"id "`
}
