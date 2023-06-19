package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(e Event) error
}

type Type int

// iota - автоитерируемая переменная
const (
	Unknown Type = iota
	Message
)

// Event - В поле Meta хранится информация, уникальная для каждого мессенджера
// Например, для ТГ нужен username, но в другом мессенджере используется userID или email.
// Этот подход позволяет сделать общую абстракцию
type Event struct {
	Type Type
	Text string
	Meta interface{} // Аналог any
}
